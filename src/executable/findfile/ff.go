package main

import (
	"crypto/md5"
	"encoding/hex"
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
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

// targets is the targets file name
var targets []string
var wg sync.WaitGroup

var verbose bool
var atomicCount atomic.Int32
var absoluteTarget atomic.Bool
var onlyDir bool = false
var printMd5 bool = false
var caseInsensitive bool = false
var relativePath bool = true
var wd string
var glob = false

var numThreads = make(chan struct{}, 50)

type conf struct {
	defaultAbs bool
}

var globalConfig *conf

const (
	absName = "ff.abs"
)

func init() {
	var err error
	wd, err = os.Getwd()
	if err != nil {
		panic(err)
	}
	wd = utilsw.Abs(wd)

	initConfig()
}

func initConfig() {
	config := utilsw.GetAllConfig()
	isAbs := config.GetOrDefault(absName, "")
	globalConfig = &conf{}
	if isAbs == "true" {
		globalConfig.defaultAbs = true
	}
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

func highlightMatch(pathStr, matchBase string) string {
	normalized := strings.ReplaceAll(pathStr, "\\", "/")
	return strings.ReplaceAll(normalized, matchBase, color.GreenString(matchBase))
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
		// already absolute path
		if len(target) > 0 && target[0] == '/' {
			matches = append(matches, target)
			absoluteTarget.Store(true)
			goto OUTER
		}
		var m []string
		if glob {
			var err error
			if caseInsensitive {
				m, err = terminalw.GlobCaseInsensitive(target, rootDir)
			} else {
				m, err = terminalw.Glob(target, rootDir)
			}
			if err != nil {
				if verbose {
					fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
				}
			}
		} else {
			for _, file := range utilsw.LsDir(rootDir, nil, nil) {
				abs := filepath.Join(rootDir, file)
				if abs == target || file == target {
					matches = append(matches, abs)
				}
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
				fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
			}
			continue
		}

		if onlyDir && !utilsw.IsDir(abs) {
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
		// Atomically increment and check if within limit
		// curCount := atomic.AddInt32(&atomicCount, 1)
		curCount := atomicCount.Add(1)
		if int(curCount) > numPrint {
			atomicCount.Add(-1) // revert increment if over limit
			break
		}
		if utilsw.IsDir(abs) && abs == filepath.Clean(abs) {
			abs = filepath.Join(abs, string(os.PathSeparator))
		}
		var toPrint string
		if !relativePath {
			toPrint = highlightMatch(abs, matchBase)
		} else {
			absStripped := strings.TrimPrefix(abs, wd)
			if absStripped != abs {
				absStripped = strings.TrimPrefix(absStripped, string(os.PathSeparator))
			}
			abs = absStripped
			toPrint = highlightMatch(absStripped, matchBase)
		}
		if verbose {
			info, err := os.Stat(match)
			if err != nil {
				fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
				continue
			}
			fileSize := info.Size()
			toPrint += "  " + parseFileSize(fileSize)
			toPrint += "  " + info.ModTime().Format("2006.01.02/15:04:05")
		}
		if printMd5 {
			b, err := os.ReadFile(abs)
			if err != nil {
				if verbose {
					fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
				}
				continue
			}
			h := md5.Sum(b)
			val := hex.EncodeToString(h[:])
			toPrint += "\t" + val
		}
		fmt.Fprintf(color.Output, "%s\n", toPrint)
		if absoluteTarget.Load() {
			os.Exit(0)
		}
	}

	// check sub directories
	subs, err := os.ReadDir(rootDir)
	if err != nil {
		if verbose {
			fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
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
	parser := terminalw.NewParser()
	parser.Int64("n", 10, "number of found results to print, -10 for short")
	parser.Bool("v", false, "if print error")
	parser.String("d", ".", "root directory for searching")
	parser.String("i", "", "ignore case")
	parser.String("ex", "", "exclude file patterns (glob )")
	parser.Bool("a", false, "list all matches (has the highest priority)")
	parser.Int("p", 4, "how many threads to use")
	parser.Bool("dir", false, "only search directories")
	parser.Bool("h", false, "print this help")
	parser.Bool("md5", false, "print md5 value")
	parser.Bool("abs", false, fmt.Sprintf("print absolute path. (%s in ~/.configW)", absName))
	parser.Bool("rel", false, "print relative path.")
	parser.Bool("glob", false, "use filepath.Glob to match")
	parser.ParseArgsCmd()

	if parser.Empty() {
		parser.PrintDefaults()
		return
	}

	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}

	printMd5 = parser.ContainsFlagStrict("md5")
	glob = parser.ContainsFlagStrict("glob")
	verboseFlag := parser.ContainsFlagStrict("v")
	rootDir := parser.GetFlagValueDefault("d", ".")
	caseInsensitive = parser.ContainsFlag("i")
	onlyDir = parser.ContainsFlagStrict("dir")
	ignores := parser.GetFlagValueDefault("ex", "")
	relativePath = !parser.ContainsFlagStrict("abs") && !globalConfig.defaultAbs
	relativePath = relativePath || parser.ContainsFlagStrict("rel")
	numPrint := parser.GetNumArgs()

	if rootDir == "~" {
		rootDir = expandTilda()
		if rootDir == "" {
			log.Fatalln("HOME is not set")
		}
	}

	if numPrint == -1 {
		numPrint, err = strconv.Atoi(parser.GetFlagValueDefault("n", "10"))

		if err != nil {
			log.Fatalln(err)
		}
	}

	if parser.ContainsFlagStrict("a") {
		numPrint = math.MaxInt32
	}

	if parser.ContainsFlagStrict("p") {
		terminalw.ChangeThreads(parser.MustGetFlagValAsInt("p"))
	}

	ignores = strings.ReplaceAll(ignores, ",", " ")
	allIgnores := strw.SplitNoEmptyPreserveQuote(ignores, ' ', `"`, false)
	for i := range allIgnores {
		temp := strings.ReplaceAll(allIgnores[i], `.`, `\.`)
		temp = strings.ReplaceAll(temp, `?`, `.`)
		temp = strings.ReplaceAll(temp, `*`, `.*`)
		allIgnores[i] = temp
	}
	// fmt.Println("allIgnores", allIgnores, parser)
	verbose = verboseFlag
	targets = parser.GetPositionalArgs(false)
	// fmt.Println("rootDir", rootDir, targets)
	allRootDirs, err := filepath.Glob(rootDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, color.RedString(err.Error()))
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
