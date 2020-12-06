// "findcont" ==> "fs.exe", stands for "find sentense"
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"

	"github.com/grewwc/go_tools/src/utilsW"
)

var target string
var wg sync.WaitGroup
var countMu sync.Mutex
var r *regexp.Regexp = nil

func colorTargetString(line string, matchedStrings []string) string {
	var result string
	for _, matchedString := range matchedStrings {
		result = strings.ReplaceAll(line, matchedString, color.RedString(matchedString))
	}
	return result
}

func checkFileFunc(filename string, fn func(target, line string) (bool, []string)) {
	file, err := os.Open(filename)
	if err != nil {
		if terminalW.Verbose {
			utilsW.Fprintln(os.Stderr, err)
		}
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineno := 0
	for scanner.Scan() {
		lineno++
		line := scanner.Text()
		matched, matchedStrings := fn(target, line)
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
	}
}

func checkFile(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		return strings.Contains(line, target), []string{target}
	})
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
	})
}

func checkFileStrict(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		return target == strings.TrimSpace(line), []string{target}
	})
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
	})
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
	})
}

func main() {
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	num := fs.Int64("n", terminalW.NumPrint, "number of found results to print")
	ext := fs.String("t", "", "what type of file to search")
	verboseFlag := fs.Bool("v", false, "if print error")
	rootDir := fs.String("d", ".", "root directory for searching")
	isReg := fs.Bool("re", false, `turn on regular expression (use "\" instead of "\\") `)
	isIgnoreCase := fs.Bool("ignore", false, "ignore upper/lower case")
	isIgnoreCaseShortcut := fs.Bool("i", false, "ignore upper/lower case (shortcut for -ignore)")
	numLevel := fs.Int("level", math.MaxInt32, `number of directory levels to search. current directory's level is 0`)
	isStrict := fs.Bool("strict", false, "find exact the same matches (after triming space)")
	extExclude := fs.String("nt", "", "check files which are not some types")
	findWord := fs.Bool("word", false, "only match the concrete word, is a shortcut for -re")
	all := fs.Bool("all", false, "shortcut for -n=-1")
	*all = *all || *fs.Bool("a", false, "shortcut for -all")          // shortcut for -all
	files := fs.String("f", "", "check only these files/directories") // this flag will override -t
	notFiles := fs.String("nf", "", "don't check these files/directories")

	fmt.Println()

	parsedResults := terminalW.ParseArgsCmd("re", "v", "ignore", "strict", "all", "word", "i")
	if parsedResults == nil {
		fs.PrintDefaults()
		return
	}
	optionalMap, args := parsedResults.Optional, parsedResults.Positional.ToStringSlice()
	optional := terminalW.MapToString(optionalMap)
	// fmt.Println("optionalMap", optionalMap, optional)
	// fmt.Println("args", args)
	// fmt.Println(optional, stringsW.SplitNoEmptyKeepQuote(optional, ' '))
	fs.Parse(stringsW.SplitNoEmptyKeepQuote(optional, ' '))

	*rootDir = filepath.ToSlash(strings.ReplaceAll(*rootDir, `\\`, `\`))
	if *num < 0 || *all {
		*num = math.MaxInt64
	}
	terminalW.NumPrint = *num
	terminalW.Verbose = *verboseFlag
	terminalW.MaxLevel = int32(*numLevel)

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

	if *findWord {
		*isReg = true
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

	if *files != "" {
		*files = strings.ReplaceAll(*files, ",", " ")
		for _, f := range stringsW.SplitNoEmpty(*files, " ") {
			terminalW.FileNamesToCheck.Add(f)
		}
		*ext = ""
		*extExclude = ""
	}

	if *notFiles != "" {
		*notFiles = strings.ReplaceAll(*notFiles, ",", "")
		for _, f := range stringsW.SplitNoEmpty(*notFiles, " ") {
			terminalW.FileNamesNOTCheck.Add(f)
			terminalW.FileNamesToCheck.Delete(f)
		}
		// because previous notFiles may make files empty
		if *files != "" && terminalW.FileNamesToCheck.Empty() {
			terminalW.FileNamesToCheck.Add(nil)
		}
	}

	*isIgnoreCase = *isIgnoreCase || *isIgnoreCaseShortcut
	if *isReg {
		task = checkFileRe
		r = regexp.MustCompile(target)
	} else if *isStrict {
		if *isIgnoreCase {
			target = strings.ToLower(strings.TrimSpace(target))
			task = checkFileStrictIgnoreCase
		} else {
			target = strings.TrimSpace(target)
			task = checkFileStrict
		}
	} else {
		if *isIgnoreCase {
			target = strings.ToLower(target)
			task = checkFileIgnoreCase
		} else {
			task = checkFile
		}
	}

	// fmt.Println("target", target)
	if *ext != "" {
		terminalW.Extensions = terminalW.FormatFileExtensions(*ext)
		terminalW.CheckExtension = true
	} else {
		terminalW.Extensions = utilsW.DefaultExtensions.ShallowCopy()
		terminalW.CheckExtension = false
	}
	if *extExclude != "" {
		// need to exclude some type of files
		excludeSet := terminalW.FormatFileExtensions(*extExclude)
		terminalW.Extensions.Subtract(*excludeSet)
		terminalW.CheckExtension = true
	}

	if *isReg && *isIgnoreCase {
		target = "(?i)" + target
		r = regexp.MustCompile(target)
	}

	fmt.Println()
	wg.Add(1)
	go terminalW.Find(*rootDir, task, &wg, 0)
	wg.Wait()

	terminalW.Once.Do(func() {
		summaryString := utilsW.Sprintf("%d matches found\n", terminalW.Count)
		utilsW.Println(strings.Repeat("-", len(summaryString)))
		matches := int64(math.Min(float64(terminalW.Count), float64(terminalW.NumPrint)))
		utilsW.Printf("%v matches found\n", matches)
	})
}
