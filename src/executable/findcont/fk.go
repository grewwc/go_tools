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

	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
)

var target string
var wg sync.WaitGroup
var countMu sync.Mutex

func checkFileFunc(filename string, fn func(target, line string) bool) {
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
		if fn(target, line) { // cannot reverse the order
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
			fmt.Printf(">> %q [%d]:  %s\n\n", filepath.ToSlash(filename), lineno,
				strings.TrimSpace(line))
		}
	}
}

func checkFile(filename string) {
	checkFileFunc(filename, func(target, line string) bool {
		return strings.Contains(line, target)
	})
}

func checkFileIgnoreCase(filename string) {
	checkFileFunc(filename, func(target, line string) bool {
		return strings.Contains(strings.ToLower(line), strings.ToLower(target))
	})
}

func checkFileStrict(filename string) {
	checkFileFunc(filename, func(target, line string) bool {
		return strings.TrimSpace(target) == strings.TrimSpace(line)
	})
}

func checkFileStrictIgnoreCase(filename string) {
	checkFileFunc(filename, func(target, line string) bool {
		return strings.ToLower(strings.TrimSpace(target)) == strings.ToLower(strings.TrimSpace(line))
	})
}

func checkFileRe(filename string) {
	checkFileFunc(filename, func(pattern, s string) bool {
		res, _ := regexp.MatchString(pattern, s)
		return res
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
	fs.BoolVar(&terminalW.CheckFileWithoutExt, "noext", false, "check file without extension")
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
		terminalW.Extensions = strings.Join(terminalW.DefaultExtensions[:], " ")
		terminalW.CheckExtension = false
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

}
