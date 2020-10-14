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
			terminalW.CountMu.Lock()
			terminalW.Count++
			if terminalW.Count > terminalW.NumPrint {
				terminalW.CountMu.Unlock()
				return
			}
			terminalW.CountMu.Unlock()
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
		return strings.TrimSpace(target) == line
	})
}

func checkFileStrictIgnoreCase(filename string) {
	checkFileFunc(filename, func(target, line string) bool {
		return strings.ToLower(strings.TrimSpace(target)) == strings.ToLower(line)
	})
}

func checkFileRe(filename string) {
	checkFileFunc(filename, func(pattern, s string) bool {
		res, _ := regexp.MatchString(pattern, s)
		return res
	})
}

func main() {
	quoteArgs := terminalW.ParseArgs("re", "v", "ignore", "-strict")
	optionalMap, args := quoteArgs.Optional, quoteArgs.Positional
	optional := terminalW.MapToString(optionalMap)
	// fmt.Println("optionalMap", optionalMap)
	// fmt.Println("args", args)

	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	num := fs.Int64("n", terminalW.NumPrint, "number of found results to print")
	ext := fs.String("t", "", "what type of file to search")
	verboseFlag := fs.Bool("v", false, "if print error")
	rootDir := fs.String("d", ".", "root directory for searching")
	isReg := fs.Bool("re", false, `if use regular expression (use "\" instead of "\\") `)
	isIgnoreCase := fs.Bool("ignore", false,
		"case sensitive or not (must use with \"-re\")")
	numLevel := fs.Int("level", math.MaxInt32,
		`how many more directory levels to search. e.g.: src/ main.go "main.go" is the level 0,
"src" is the level 1`)
	isStrict := fs.Bool("strict", false, "find exact the same matches (after triming space)")

	fs.Parse(stringsW.SplitNoEmptyKeepQuote(optional, ' '))

	*rootDir = filepath.ToSlash(strings.ReplaceAll(*rootDir, `\\`, `\`))
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
	} else {
		terminalW.Extensions = strings.Join(terminalW.DefaultExtensions[:], " ")
	}
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
	summaryString := fmt.Sprintf("%d matches found\n", terminalW.Count)
	fmt.Println(strings.Repeat("-", len(summaryString)))
	fmt.Printf("%v matches found\n", math.Min(float64(terminalW.Count), float64(terminalW.NumPrint)))
}
