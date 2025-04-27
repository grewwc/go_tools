// "findcont" ==> "fs.exe", stands for "find sentense"
package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

var target string
var wg sync.WaitGroup
var r *regexp.Regexp = nil
var numLines int = 1

const (
	MAX_LEN = 128
)

func colorTargetString(line string, matchedStrings []string) string {
	var result string
	for _, matchedString := range matchedStrings {
		result = strings.ReplaceAll(line, matchedString, color.RedString(matchedString))
	}
	return result
}

func checkFileFunc(filename string, fn func(target, line string) (bool, []string), numLines int) {
	file, err := os.Open(filename)
	if err != nil {
		if terminalw.Verbose {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	defer file.Close()
	var matched bool
	var matchedStrings []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 256), 1024*1024*1024)
	lineno := 0
	lineCnt := 1
	var line string
	for scanner.Scan() {
		lineno++
		if lineCnt == 1 {
			line = scanner.Text()
		}
		for lineCnt < numLines && scanner.Scan() {
			// no, _ := utf8.DecodeLastRuneInString(line)
			// if no < 256 && no != ' ' {
			// 	sep = " "
			// }
			line += scanner.Text()
			// fmt.Println("here===> ", line)
			lineno++
			lineCnt++
			matched, matchedStrings = fn(target, line)
			if matched {
				goto noMatchNeed
			} else {
				line = strw.SubStringQuiet(line, len(line)-len(target), len(line))
			}
		}
		matched, matchedStrings = fn(target, line)
	noMatchNeed:
		if matched { // cannot reverse the order
			atomic.AddInt64(&terminalw.Count, 1)
			if terminalw.Count > terminalw.NumPrint {
				return
			}
			filename, err = filepath.Abs(filename)
			if err != nil {
				if terminalw.Verbose {
					fmt.Fprintln(os.Stderr, err)
				}
				return
			}
			filename = filepath.ToSlash(filename)
			dir := filepath.Dir(filename)
			base := filepath.Base(filename)
			line := strw.SubStringQuiet(line, 0, MAX_LEN)
			fmt.Fprintf(color.Output, "%s \"%s%c%s\" [%d]:  %s\n\n", color.GreenString(">>"),
				dir, filepath.Separator, color.YellowString(base), lineno,
				colorTargetString(strings.TrimSpace(line), matchedStrings))
		}
		line = ""
		lineCnt = 1
		if scanner.Err() != nil {
			log.Println(scanner.Err())
		}
	}
}

func checkFile(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		return strings.Contains(line, target), []string{target}
	}, numLines)
}

func checkFileIgnoreCase(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		// only support English
		// undefined behavior for other languages
		prevIdx, idx := 0, 0
		result := make([]string, 1)
		count := 0
		for {
			idx = strings.Index(strings.ToLower(line[prevIdx:]), target)
			if idx == -1 {
				return count != 0, result
			}
			idx += prevIdx
			count++
			result = append(result, line[idx:idx+len(target)])
			idx += len(target)
			prevIdx = idx
		}
	}, numLines)
}

func checkFileStrict(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		return target == strings.TrimSpace(line), []string{target}
	}, numLines)
}

func checkFileStrictIgnoreCase(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		line = strings.TrimSpace(line)
		result := make([]string, 1)
		count := 0
		for idx := range line {
			idx = strings.Index(strings.ToLower(line), target)
			if idx == -1 {
				return count != 0, result
			}
			count++
			result = append(result, line[idx:idx+len(target)])
		}
		// should not reach here forever
		return false, nil
	}, numLines)
}

func checkFileRe(filename string) {
	checkFileFunc(filename, func(pattern, s string) (bool, []string) {
		result := r.FindAllString(s, -1)
		if result == nil {
			return false, nil
		}
		if len(result) == 0 {
			return false, result
		}
		return true, result
	}, numLines)
}

func main() {
	var num int64 = 5
	var isReg = false
	var isIgnoreCase = false
	parser := terminalw.NewParser()
	parser.Int64("n", terminalw.NumPrint, "number of found results to print")
	parser.String("t", "", "what type of file to search")
	parser.Bool("v", false, "if print error")
	parser.String("d", ".", "root directory for searching")
	parser.Bool("re", false, `turn on regular expression (use "\" instead of "\\") `)
	parser.Bool("ignore", false, "ignore upper/lower case")
	parser.Bool("i", false, "ignore upper/lower case (shortcut for -ignore)")
	parser.Int("level", math.MaxInt32, `number of directory levels to search. current directory's level is 0`)
	parser.Bool("strict", false, "find exact the same matches (after triming space)")
	parser.String("nt", "", "check files which are not some types")
	parser.Bool("word", false, "only match the concrete word, is a shortcut for -re")
	parser.Bool("all", false, "shortcut for -n=-1")
	parser.Bool("a", false, "shortcut for -all")
	parser.String("f", "", "check only these files/directories") // this flag will override -t
	parser.String("nf", "", "don't check these files/directories")
	parser.Int("l", 1, "how many lines more read to match")
	parser.Int("p", 4, "how many threads to use")
	parser.Bool("h", false, "print help info")

	fmt.Println()

	parser.ParseArgsCmd("re", "v", "ignore", "strict", "all", "word", "i", "a", "h")
	// fmt.Println("here", parser)
	if parser.Empty() || parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}

	args := parser.GetPositionalArgs(true)
	if parser.GetNumArgs() != -1 {
		num = int64(parser.GetNumArgs())
	}

	ext := parser.GetFlagValueDefault("t", "")
	rootDir := filepath.ToSlash(strings.ReplaceAll(parser.GetFlagValueDefault("d", "."), `\\`, `\`))

	// fmt.Println(parser)
	all := parser.ContainsFlagStrict("all") || parser.ContainsFlagStrict("a")
	if num < 0 || all {
		num = math.MaxInt64
	}
	terminalw.NumPrint = num
	terminalw.Verbose = parser.ContainsFlagStrict("v")
	temp, err := strconv.Atoi(parser.GetFlagValueDefault("level", strconv.Itoa(math.MaxInt32)))
	if err != nil {
		log.Fatalln(err)
	}
	terminalw.MaxLevel = int32(temp)

	// below the main thing is to define the task
	var task func(string)

	// fmt.Println(terminalw.Extensions)
	switch len(args) {
	case 1:
		target = args[0]
	default:
		parser.PrintDefaults()
		return
	}
	target = strings.ReplaceAll(target, `\\`, `\`)

	if parser.ContainsFlagStrict("re") {
		isReg = true
	}

	if parser.ContainsFlagStrict("word") {
		isReg = true
		wordPattern := regexp.MustCompile(`\w+`)
		if !wordPattern.MatchString(target) {
			// fmt.Println("here", target)
			fmt.Println("You should pass in a word if set \"-word\" option")
			parser.PrintDefaults()
			os.Exit(1)
		}
		target = fmt.Sprintf("\\b%s\\b", target)
		r = regexp.MustCompile(target)
	}

	extExclude := parser.GetFlagValueDefault("nt", "")
	files := parser.GetFlagValueDefault("f", "")
	notFiles := parser.GetFlagValueDefault("nf", "")

	if files != "" {
		files = strings.ReplaceAll(files, ",", " ")
		for _, f := range strw.SplitNoEmpty(files, " ") {
			terminalw.FileNamesToCheck.Add(f)
		}
		ext = ""
		extExclude = ""
	}

	if parser.ContainsFlagStrict("l") {
		numLines = parser.MustGetFlagValAsInt("l")
	}

	if parser.ContainsFlagStrict("p") {
		res := parser.MustGetFlagValAsInt("p")
		terminalw.ChangeThreads(res)
	}

	if notFiles != "" {
		notFiles = strings.ReplaceAll(notFiles, ",", "")
		for _, f := range strw.SplitNoEmpty(notFiles, " ") {
			terminalw.FileNamesNOTCheck.Add(f)
			terminalw.FileNamesToCheck.Delete(f)
		}
		// because previous notFiles may make files empty
		if files != "" && terminalw.FileNamesToCheck.Empty() {
			terminalw.FileNamesToCheck.Add(nil)
		}
	}

	isIgnoreCase = parser.ContainsFlagStrict("ignore") || parser.ContainsFlagStrict("i")

	if isReg {
		task = checkFileRe
		r = regexp.MustCompile(target)
	} else if parser.ContainsFlagStrict("strict") {
		if isIgnoreCase {
			target = strings.ToLower(strings.TrimSpace(target))
			task = checkFileStrictIgnoreCase
		} else {
			target = strings.TrimSpace(target)
			task = checkFileStrict
		}
	} else {
		if isIgnoreCase {
			target = strings.ToLower(target)
			task = checkFileIgnoreCase
		} else {
			task = checkFile
		}
	}

	if ext != "" {
		terminalw.Extensions = terminalw.FormatFileExtensions(ext)
		terminalw.CheckExtension = true
	} else {
		terminalw.Extensions = utilsw.DefaultExtensions.ShallowCopy()
		terminalw.CheckExtension = false
	}
	if extExclude != "" {
		// need to exclude some type of files
		excludeSet := terminalw.FormatFileExtensions(extExclude)
		terminalw.Extensions.Subtract(*excludeSet)
		terminalw.CheckExtension = true
	}

	if isReg && isIgnoreCase {
		target = "(?i)" + target
		r = regexp.MustCompile(target)
	}

	// fmt.Println(numLines)
	fmt.Println()
	wg.Add(1)
	go terminalw.Find(rootDir, task, &wg, 0)
	wg.Wait()

	terminalw.Once.Do(func() {
		summaryString := fmt.Sprintf("%d matches found\n", terminalw.Count)
		fmt.Println(strings.Repeat("-", len(summaryString)))
		matches := int64(math.Min(float64(terminalw.Count), float64(terminalw.NumPrint)))
		fmt.Printf("%v matches found\n", matches)
	})
}
