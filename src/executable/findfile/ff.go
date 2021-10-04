package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

// target is the target file name
var target string
var wg sync.WaitGroup

var numPrint int64

var verbose bool
var ignores string
var count int64

var numThreads = make(chan struct{}, 50)

var mu = sync.Mutex{}

func findFile(rootDir string) {
	numThreads <- struct{}{}
	defer func() { <-numThreads }()
	defer wg.Done()

	matches, err := terminalW.Glob(target, rootDir)
	if err != nil {
		if verbose {
			utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
		}
		return
	}
OUTER:
	for _, match := range matches {
		mu.Lock()
		if atomic.LoadInt64(&count) >= numPrint {
			mu.Unlock()
			return
		}
		mu.Unlock()
		abs, err := filepath.Abs(match)
		if err != nil {
			if verbose {
				utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
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
		match = filepath.Base(match)
		utilsW.Fprintf(color.Output, "%s %s\n", color.GreenString(">>"),
			strings.ReplaceAll(strings.ReplaceAll(abs, "\\", "/"), match, color.GreenString(match)))
	}

	// check sub directories
	subs, err := ioutil.ReadDir(rootDir)
	if err != nil {
		if verbose {
			utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
		}
		return
	}

	for _, sub := range subs {
		if sub.IsDir() {
			wg.Add(1)
			go findFile(path.Join(rootDir, sub.Name()))
		}
	}
}

func main() {
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.Int64Var(&numPrint, "n", 10, "number of found results to print")
	verboseFlag := fs.Bool("v", false, "if print error")
	rootDir := fs.String("d", ".", "root directory for searching")
	fs.StringVar(&ignores, "i", "", "ignores some file pattern")

	results := terminalW.ParseArgsCmd("v")
	if results.ContainsFlagStrict("v") {
		*verboseFlag = true
	}

	*rootDir = results.GetFlagValueDefault("d", ".")
	ignores = results.GetFlagValueDefault("i", "")
	numPrint, err := strconv.ParseInt(results.GetFlagValueDefault("n", "10"), 10, 64)
	if err != nil {
		log.Fatalln(err)
	}

	res := terminalW.ParseArgsCmd(strings.Join(terminalW.AddQuote(os.Args[1:]), " "))

	if res == nil {
		fs.PrintDefaults()
		return
	}
	optionalMap, args := res.Optional, res.Positional.ToStringSlice()
	optional := terminalW.MapToString(optionalMap)

	fs.Parse(stringsW.SplitNoEmptyKeepQuote(optional, ' '))

	if res == nil {
		fs.PrintDefaults()
		return
	}

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
	// fmt.Println("rootDir", *rootDir)
	allRootDirs, err := filepath.Glob(*rootDir)
	if err != nil {
		utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
		return
	}
	for _, dir := range allRootDirs {
		wg.Add(1)
		go findFile(dir)
	}
	wg.Wait()
	summaryString := fmt.Sprintf("%d matches found\n", count)
	fmt.Println(strings.Repeat("-", len(summaryString)))
	fmt.Printf("%v matches found\n", math.Min(float64(count), float64(numPrint)))
}
