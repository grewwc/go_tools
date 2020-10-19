package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
)

// target is the target file name
var target string
var wg sync.WaitGroup

var numPrint int64

var verbose bool
var ignores string
var count int64

var numThreads = make(chan struct{}, 5000)

func checkError(err error) {
	if verbose && err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
}

func findFile(rootDir string) {
	numThreads <- struct{}{}
	defer func() { <-numThreads }()
	defer wg.Done()

	matches, err := terminalW.Glob(target, rootDir)
	checkError(err)
OUTER:
	for _, match := range matches {
		if atomic.LoadInt64(&count) >= numPrint {
			return
		}
		abs, err := filepath.Abs(match)
		if err != nil {
			if verbose {
				fmt.Fprintln(os.Stderr, err)
			}
			continue
		}

		allIgnores := stringsW.SplitNoEmptyKeepQuote(ignores, ' ')
		for _, toIgnore := range allIgnores {
			if strings.Contains(abs, toIgnore) {
				continue OUTER
			}
		}
		atomic.AddInt64(&count, 1)
		fmt.Printf(">> %q\n", strings.ReplaceAll(abs, "\\", "/"))
	}

	// check sub directories
	subs, err := ioutil.ReadDir(rootDir)
	checkError(err)

	for _, sub := range subs {
		if sub.IsDir() {
			wg.Add(1)
			go findFile(path.Join(rootDir, sub.Name()))
		}
	}
}

func main() {
	res := terminalW.ParseArgsCmd(strings.Join(terminalW.AddQuote(os.Args[1:]), " "))
	optionalMap, args := res.Optional, res.Positional
	optional := terminalW.MapToString(optionalMap)
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.Int64Var(&numPrint, "n", 10, "number of found results to print")
	verboseFlag := fs.Bool("v", false, "if print error")
	rootDir := fs.String("d", ".", "root directory for searching")
	fs.StringVar(&ignores, "i", "", "ignores some file pattern")
	fs.Parse(stringsW.SplitNoEmptyKeepQuote(optional, ' '))

	ignores = strings.ReplaceAll(ignores, ",", " ")
	verbose = *verboseFlag
	switch len(args) {
	case 1:
		target = args[0]
	default:
		fs.PrintDefaults()
		return
	}
	fmt.Println()
	wg.Add(1)
	go findFile(*rootDir)
	wg.Wait()
	summaryString := fmt.Sprintf("%d matches found\n", count)
	fmt.Println(strings.Repeat("-", len(summaryString)))
	fmt.Printf("%v matches found\n", math.Min(float64(count), float64(numPrint)))
}
