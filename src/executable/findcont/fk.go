// "findcont" ==> "fs.exe", stands for "find sentense"
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"

	"github.com/grewwc/go_tools/src/utilsW"
)

var target string
var wg sync.WaitGroup
var countMu sync.Mutex
var r *regexp.Regexp = nil
var numLines int = 1

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
		if terminalW.Verbose {
			utilsW.Fprintln(os.Stderr, err)
		}
		return
	}
	defer file.Close()
	var matched bool
	var matchedStrings []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 100), 1024*1024*16)
	lineno := 0
	lineCnt := 1
	var line string
	for scanner.Scan() {
		lineno++
		if lineCnt == 1 {
			line = scanner.Text()
		}
		for lineCnt < numLines && scanner.Scan() {
			var sep string
			no, _ := utf8.DecodeLastRuneInString(line)
			if no < 256 && no != ' ' {
				sep = " "
			}
			line += sep + strings.TrimSpace(scanner.Text())
			// fmt.Println("here===> ", line)
			lineno++
			lineCnt++
			matched, matchedStrings = fn(target, line)
			if matched {
				goto noMatchNeed
			}
		}
		matched, matchedStrings = fn(target, line)
	noMatchNeed:
		if matched { // cannot reverse the order
			countMu.Lock()
			terminalW.Count++
			if terminalW.Count > terminalW.NumPrint {
				countMu.Unlock()
				return
			}
			countMu.Unlock()
			filename, err = filepath.Abs(filename)
			if err != nil {
				if terminalW.Verbose {
					utilsW.Fprintln(os.Stderr, err)
				}
				return
			}
			filename = filepath.ToSlash(filename)
			dir := filepath.Dir(filename)
			base := filepath.Base(filename)

			utilsW.Fprintf(color.Output, "%s \"%s%c%s\" [%d]:  %s\n\n", color.GreenString(">>"),
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
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.Int64("n", terminalW.NumPrint, "number of found results to print")
	fs.String("t", "", "what type of file to search")
	fs.Bool("v", false, "if print error")
	fs.String("d", ".", "root directory for searching")
	fs.Bool("re", false, `turn on regular expression (use "\" instead of "\\") `)
	fs.Bool("ignore", false, "ignore upper/lower case")
	fs.Bool("i", false, "ignore upper/lower case (shortcut for -ignore)")
	fs.Int("level", math.MaxInt32, `number of directory levels to search. current directory's level is 0`)
	fs.Bool("strict", false, "find exact the same matches (after triming space)")
	fs.String("nt", "", "check files which are not some types")
	fs.Bool("word", false, "only match the concrete word, is a shortcut for -re")
	fs.Bool("all", false, "shortcut for -n=-1")
	fs.Bool("a", false, "shortcut for -all")
	fs.String("f", "", "check only these files/directories") // this flag will override -t
	fs.String("nf", "", "don't check these files/directories")
	fs.Int("l", 1, "how many lines more read to match")
	fs.Int("p", 4, "how many threads to use")

	fmt.Println()

	parsedResults := terminalW.ParseArgsCmd("re", "v", "ignore", "strict", "all", "word", "i", "a")
	// fmt.Println("here", parsedResults)
	if parsedResults == nil {
		fs.PrintDefaults()
		return
	}

	optionalMap, args := parsedResults.Optional, parsedResults.Positional.ToStringSlice()
	optional := terminalW.MapToString(optionalMap)
	if parsedResults.GetNumArgs() != -1 {
		num = int64(parsedResults.GetNumArgs())
		r := regexp.MustCompile("-\\d+")
		optional = r.ReplaceAllString(optional, "")
	}

	ext := parsedResults.GetFlagValueDefault("t", "")
	rootDir := filepath.ToSlash(strings.ReplaceAll(parsedResults.GetFlagValueDefault("d", "."), `\\`, `\`))

	// fmt.Println(parsedResults)
	all := parsedResults.ContainsFlagStrict("all") || parsedResults.ContainsFlagStrict("a")
	if num < 0 || all {
		num = math.MaxInt64
	}
	terminalW.NumPrint = num
	terminalW.Verbose = parsedResults.ContainsFlagStrict("v")
	temp, err := strconv.Atoi(parsedResults.GetFlagValueDefault("level", strconv.Itoa(math.MaxInt32)))
	if err != nil {
		log.Fatalln(err)
	}
	terminalW.MaxLevel = int32(temp)

	// below the main thing is to define the task
	var task func(string)

	// fmt.Println(terminalW.Extensions)
	switch len(args) {
	case 1:
		target = args[0]
	default:
		fs.PrintDefaults()
		return
	}
	target = strings.ReplaceAll(target, `\\`, `\`)

	if parsedResults.ContainsFlagStrict("re") {
		isReg = true
	}

	if parsedResults.ContainsFlagStrict("word") {
		isReg = true
		wordPattern := regexp.MustCompile("\\w+")
		if !wordPattern.MatchString(target) {
			// fmt.Println("here", target)
			fmt.Println("You should pass in a word if set \"-word\" option")
			fs.PrintDefaults()
			os.Exit(1)
		}
		target = fmt.Sprintf("\\b%s\\b", target)
		r = regexp.MustCompile(target)
	}

	extExclude := parsedResults.GetFlagValueDefault("nt", "")
	files := parsedResults.GetFlagValueDefault("f", "")
	notFiles := parsedResults.GetFlagValueDefault("nf", "")

	if files != "" {
		files = strings.ReplaceAll(files, ",", " ")
		for _, f := range stringsW.SplitNoEmpty(files, " ") {
			terminalW.FileNamesToCheck.Add(f)
		}
		ext = ""
		extExclude = ""
	}

	if parsedResults.ContainsFlagStrict("l") {
		numLines = parsedResults.MustGetFlagValAsInt("l")
	}

	if parsedResults.ContainsFlagStrict("p") {
		res := parsedResults.MustGetFlagValAsInt("p")
		terminalW.ChangeThreads(res)
	}

	if notFiles != "" {
		notFiles = strings.ReplaceAll(notFiles, ",", "")
		for _, f := range stringsW.SplitNoEmpty(notFiles, " ") {
			terminalW.FileNamesNOTCheck.Add(f)
			terminalW.FileNamesToCheck.Delete(f)
		}
		// because previous notFiles may make files empty
		if files != "" && terminalW.FileNamesToCheck.Empty() {
			terminalW.FileNamesToCheck.Add(nil)
		}
	}

	isIgnoreCase = parsedResults.ContainsFlagStrict("ignore") || parsedResults.ContainsFlagStrict("i")

	if isReg {
		task = checkFileRe
		r = regexp.MustCompile(target)
	} else if parsedResults.ContainsFlagStrict("strict") {
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
		terminalW.Extensions = terminalW.FormatFileExtensions(ext)
		terminalW.CheckExtension = true
	} else {
		terminalW.Extensions = utilsW.DefaultExtensions.ShallowCopy()
		terminalW.CheckExtension = false
	}
	if extExclude != "" {
		// need to exclude some type of files
		excludeSet := terminalW.FormatFileExtensions(extExclude)
		terminalW.Extensions.Subtract(*excludeSet)
		terminalW.CheckExtension = true
	}

	if isReg && isIgnoreCase {
		target = "(?i)" + target
		r = regexp.MustCompile(target)
	}

	// fmt.Println(numLines)
	fmt.Println()
	wg.Add(1)
	go terminalW.Find(rootDir, task, &wg, 0)
	wg.Wait()

	terminalW.Once.Do(func() {
		summaryString := utilsW.Sprintf("%d matches found\n", terminalW.Count)
		utilsW.Println(strings.Repeat("-", len(summaryString)))
		matches := int64(math.Min(float64(terminalW.Count), float64(terminalW.NumPrint)))
		utilsW.Printf("%v matches found\n", matches)
	})
}
