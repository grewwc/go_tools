package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

// targets is the targets file name
var targets []string
var wg sync.WaitGroup

var verbose bool
var atomicCount atomic.Int32
var onlyDir bool = false
var printMd5 bool = false
var caseInsensitive bool = false
var relativePath bool = true
var wd string

var numThreads = make(chan struct{}, 50)

func init() {
	var err error
	wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	wd = utilsW.Abs(wd)
}

func expandTilda() string {
	return os.Getenv("HOME")
}

func parseFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2fK", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fM", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%.2fG", float64(size)/1024/1024/1024)
	}
}

func findFile(rootDir string, numPrint int, allIgnores []string) {
	numThreads <- struct{}{}
	defer func() { <-numThreads }()
	defer wg.Done()

	if int(atomicCount.Load()) >= numPrint {
		return
	}

	var matches []string
	for _, target := range targets {
		var m []string
		var err error
		if caseInsensitive {
			m, err = terminalW.GlobCaseInsensitive(target, rootDir)
		} else {
			m, err = terminalW.Glob(target, rootDir)
		}
		if err != nil {
			if verbose {
				utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
			}
		}
		if len(m) == 0 {
			continue
		}
		matches = append(matches, m...)
	}
OUTER:
	for _, match := range matches {

		abs, err := filepath.Abs(match)
		if err != nil {
			if verbose {
				utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
			}
			continue
		}

		if onlyDir && !utilsW.IsDir(abs) {
			continue
		}

		for _, toIgnore := range allIgnores {
			// fmt.Println("matching ", toIgnore, abs)
			if match, _ := regexp.MatchString(toIgnore, filepath.ToSlash(abs)); match {
				// fmt.Println("here", toIgnore)
				continue OUTER
			}
		}
		matchBase := filepath.Base(match)
		if int(atomicCount.Load()) < numPrint {
			if utilsW.IsDir(abs) && !strings.HasSuffix(abs, "/") {
				abs += "/"
			}
			var toPrint string
			if !relativePath {
				toPrint = strings.ReplaceAll(strings.ReplaceAll(abs, "\\", "/"), matchBase, color.GreenString(matchBase))
			} else {
				absStripped := stringsW.StripPrefix(abs, wd)
				if absStripped != abs {
					absStripped = stringsW.StripPrefix(absStripped, "/")
				}
				abs = absStripped
				toPrint = strings.ReplaceAll(strings.ReplaceAll(absStripped, "\\", "/"), matchBase, color.GreenString(matchBase))
			}
			if verbose {
				info, err := os.Stat(match)
				fileSize := info.Size()
				if err != nil {
					utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
				}
				toPrint += "  " + parseFileSize(fileSize)
				toPrint += "  " + info.ModTime().Format("2006/01/02/15:04:05")
			}
			if printMd5 {
				b, err := os.ReadFile(abs)
				if err != nil {
					if verbose {
						utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
					}
					continue
				}
				h := md5.Sum(b)
				val := hex.EncodeToString(h[:])
				toPrint += "\t" + val
			}
			utilsW.Fprintf(color.Output, "%s\n", toPrint)
			atomicCount.Add(1)
		}
	}

	// check sub directories
	subs, err := os.ReadDir(rootDir)
	if err != nil {
		if verbose {
			utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
		}
		return
	}

	for _, sub := range subs {
		if sub.IsDir() {
			wg.Add(1)
			go findFile(path.Join(rootDir, sub.Name()), numPrint, allIgnores)
		}
	}
}

func main() {
	var err error
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.Int64("n", 10, "number of found results to print, -10 for short")
	fs.Bool("v", false, "if print error")
	fs.String("d", ".", "root directory for searching")
	fs.String("i", "", "ignore case")
	fs.String("ex", "", "exclude file patterns (glob )")
	fs.Bool("a", false, "list all matches (has the highest priority)")
	fs.Int("p", 4, "how many threads to use")
	fs.Bool("dir", false, "only search directories")
	fs.Bool("h", false, "print this help")
	fs.Bool("md5", false, "print md5 value")
	fs.Bool("abs", false, "print absolute path")
	results := terminalW.ParseArgsCmd("v", "a", "dir", "h", "md5", "abs")

	if results == nil {
		fs.PrintDefaults()
		return
	}

	if results.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}

	if results.ContainsFlagStrict("md5") {
		printMd5 = true
	}

	verboseFlag := results.ContainsFlagStrict("v")

	rootDir := results.GetFlagValueDefault("d", ".")
	if rootDir == "~" {
		rootDir = expandTilda()
		if rootDir == "" {
			log.Fatalln("HOME is not set")
		}
	}
	ignores := results.GetFlagValueDefault("ex", "")
	caseInsensitive = results.ContainsFlag("i")

	if results.ContainsFlagStrict("dir") {
		onlyDir = true
	}

	relativePath = !results.ContainsFlagStrict("abs")

	numPrint := results.GetNumArgs()
	if numPrint == -1 {
		numPrint, err = strconv.Atoi(results.GetFlagValueDefault("n", "10"))

		if err != nil {
			log.Fatalln(err)
		}
	}

	if results.ContainsFlagStrict("a") {
		numPrint = math.MaxInt32
	}

	if results.ContainsFlagStrict("p") {
		terminalW.ChangeThreads(results.MustGetFlagValAsInt("p"))
	}

	ignores = strings.ReplaceAll(ignores, ",", " ")
	allIgnores := stringsW.SplitNoEmptyKeepQuote(ignores, ' ')
	for i := range allIgnores {
		temp := strings.ReplaceAll(allIgnores[i], `.`, `\.`)
		temp = strings.ReplaceAll(temp, `?`, `.`)
		temp = strings.ReplaceAll(temp, `*`, `.*`)
		allIgnores[i] = temp
	}
	// fmt.Println("allIgnores", allIgnores, results)
	verbose = verboseFlag
	targets = results.Positional.ToStringSlice()

	// fmt.Println("rootDir", *rootDir)
	allRootDirs, err := filepath.Glob(rootDir)
	if err != nil {
		utilsW.Fprintln(os.Stderr, color.RedString(err.Error()))
		return
	}
	for _, dir := range allRootDirs {
		wg.Add(1)
		go findFile(dir, numPrint, allIgnores)
	}
	wg.Wait()
	count := atomicCount.Load()
	summaryString := fmt.Sprintf("%d matches found\n", count)
	if count > 1 && verboseFlag {
		fmt.Println(strings.Repeat("-", len(summaryString)))
		fmt.Printf("%v matches found\n", math.Min(float64(count), float64(numPrint)))
	}
}
