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
)

var target string
var wg sync.WaitGroup
var countMu sync.Mutex

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
			fmt.Fprintln(os.Stderr, err)
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
					fmt.Fprintln(os.Stderr, err)
				}
				return
			}
			filename = filepath.ToSlash(filename)
			dir := filepath.Dir(filename)
			base := filepath.Base(filename)
			fmt.Fprintf(color.Output, "%s \"%s%c%s\" [%d]:  %s\n\n", color.GreenString(">>"),
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
		idx := 0
		result := make([]string, 1)
		targetLower := strings.ToLower(target)
		count := 0
		for {
			idx = strings.Index(strings.ToLower(line[idx:]), targetLower)
			if idx == -1 {
				return count != 0, result
			}
			count++
			result = append(result, line[idx:idx+len(targetLower)])
			idx += len(targetLower)
		}
	})
}

func checkFileStrict(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		return strings.TrimSpace(target) == strings.TrimSpace(line), []string{target}
	})
}

func checkFileStrictIgnoreCase(filename string) {
	checkFileFunc(filename, func(target, line string) (bool, []string) {
		targetLower := strings.ToLower(strings.TrimSpace(target))
		line = strings.TrimSpace(line)
		result := make([]string, 1)
		count := 0
		for idx := range line {
			idx = strings.Index(strings.ToLower(line), targetLower)
			if idx == -1 {
				return count != 0, result
			}
			count++
			result = append(result, line[idx:idx+len(targetLower)])
		}
		// should not reach here forever
		return false, nil
	})
}

func checkFileRe(filename string) {
	checkFileFunc(filename, func(pattern, s string) (bool, []string) {
		r := regexp.MustCompile(pattern)
		result := r.FindAllString(s, -1)
		if result == nil {
			return false, nil
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
	numLevel := fs.Int("level", math.MaxInt32, `number of directory levels to search. current directory's level is 0`)
	isStrict := fs.Bool("strict", false, "find exact the same matches (after triming space)")
	extExclude := fs.String("nt", "", "check files which are not some types")
	fmt.Println()

	parsedResults := terminalW.ParseArgsCmd("re", "v", "ignore", "strict")
	if parsedResults == nil {
		fs.PrintDefaults()
		return
	}
	optionalMap, args := parsedResults.Optional, parsedResults.Positional
	optional := terminalW.MapToString(optionalMap)
	// fmt.Println("optionalMap", optionalMap)
	// fmt.Println("args", args)
	// fmt.Println(optional, stringsW.SplitNoEmptyKeepQuote(optional, ' '))
	fs.Parse(stringsW.SplitNoEmptyKeepQuote(optional, ' '))

	*rootDir = filepath.ToSlash(strings.ReplaceAll(*rootDir, `\\`, `\`))
	if *num < 0 {
		*num = math.MaxInt64
	}
	terminalW.NumPrint = *num
	terminalW.Verbose = *verboseFlag
	terminalW.MaxLevel = int32(*numLevel)

	var task func(string)
	if *isReg {
		task = checkFileRe
	} else if *isStrict {
		if *isIgnoreCase {
			task = checkFileStrictIgnoreCase
		} else {
			task = checkFileStrict
		}
	} else {
		if *isIgnoreCase {
			task = checkFileIgnoreCase
		} else {
			task = checkFile
		}
	}
	if *ext != "" {
		terminalW.Extensions = terminalW.FormatFileExtensions(*ext)
		terminalW.CheckExtension = true
	} else {
		terminalW.Extensions = terminalW.DefaultExtensions.ShallowCopy()
		terminalW.CheckExtension = false
	}
	if *extExclude != "" {
		// need to exclude some type of files
		excludeSet := terminalW.FormatFileExtensions(*extExclude)
		terminalW.Extensions.Subtract(*excludeSet)
		terminalW.CheckExtension = true
	}

	// fmt.Println(terminalW.Extensions)
	switch len(args) {
	case 1:
		target = args[0]
	default:
		fs.PrintDefaults()
		return
	}

	target = strings.ReplaceAll(target, `\\`, `\`)
	if *isReg && *isIgnoreCase {
		target = "(?i)" + target
	}
	fmt.Println()
	wg.Add(1)
	go terminalW.Find(*rootDir, task, &wg, 0)
	wg.Wait()

	terminalW.Once.Do(func() {
		summaryString := fmt.Sprintf("%d matches found\n", terminalW.Count)
		fmt.Println(strings.Repeat("-", len(summaryString)))
		matches := int64(math.Min(float64(terminalW.Count), float64(terminalW.NumPrint)))
		fmt.Printf("%v matches found\n", matches)
	})

}
