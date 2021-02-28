package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	_lsW "github.com/grewwc/go_tools/src/executable/ls/utils"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"github.com/grewwc/go_tools/src/windowsW"
	"golang.org/x/sys/windows"
)

var w int
var all bool

var errMsgs = containerW.NewQueue()

func init() {
	windowsW.EnableVirtualTerminal()
	var info windows.ConsoleScreenBufferInfo
	stdout := windows.Handle(os.Stdout.Fd())

	windows.GetConsoleScreenBufferInfo(stdout, &info)
	w = int(info.Size.X)
}

func formatFileStat(filename string) string {
	stat, err := os.Stat(filename)
	if err != nil {
		errMsgs.Enqueue(fmt.Sprintf("error getting stat of file: %q\n", filename))
		return ""
		// os.Exit(1)
	}
	modTime := stat.ModTime()
	modTimeStr := fmt.Sprintf("    %04d/%02d/%02d  %02d:%02d", modTime.Year(), int(modTime.Month()), modTime.Day(), modTime.Hour(), modTime.Minute())
	var sizeStr string
	if !utilsW.IsDir(filename) {
		sizeStr = stringsW.FormatInt64(stat.Size())
	} else {
		dirSize, err := utilsW.GetDirSize(filename)
		if err != nil {
			errMsgs.Enqueue(fmt.Sprintf("error getting size of directory: %q\n", filename))
			return ""
			// log.Printf("error getting size of directory: %q\n", filename)
			// os.Exit(1)
		}
		sizeStr = stringsW.FormatInt64(dirSize)
	}
	if utilsW.IsDir(filename) {
		filename = color.HiBlueString(filename + "/")
	}

	return fmt.Sprintf("%s\t%s\t%s", modTimeStr, sizeStr, filepath.ToSlash(filename))
}

func printErrors() {
	if errMsgs.Empty() {
		return
	}
	fmt.Println()
	fmt.Println("Errors:")
	count := 1
	for !errMsgs.Empty() {
		fmt.Printf("  %d: %s\n", count, color.RedString(errMsgs.Dequeue().(string)))
		count++
	}
}

func processSingleDir(rootDir string, fileSlice []string, long bool, sortType int,
	coloredStrings *containerW.Set) string {
	// if sortType != _lsW.Unsort
	// sort the fileSlice
	if sortType != _lsW.Unsort {
		fileSlice = _lsW.SortByModifiedDate(fileSlice, sortType)
	}
	// fmt.Println("here", fileSlice, sortType)
	files := ""
	for _, file := range fileSlice {
		file = filepath.Join(rootDir, file)
		file = filepath.ToSlash(file)
		if !all && filepath.Base(file)[0] == '.' {
			continue
		}
		if long {
			line := formatFileStat(file)
			if line != "" {
				files += line + "\x01\n"
			}
			continue
		}
		if utilsW.IsDir(file) {
			file += "/"
			coloredStrings.Add(file)
		}
		if strings.Contains(file, " ") {
			file = fmt.Sprintf("\"%s\"", file)
			coloredStrings.Add(file)
			file = strings.ReplaceAll(file, " ", "\x00")
		}
		files += file
		files += " "
	}
	return files
}

func main() {
	var files string
	var sortType int = _lsW.Unsort
	var l bool
	var numFileToPrint int = math.MaxInt32

	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.Bool("l", false, "show more information")
	fs.Bool("a", false, "list hidden file")
	fs.Bool("t", false, "sort files by last modified date")
	fs.Bool("rt", false, "sort files by earlist modified date")
	fs.Bool("h", false, "print help information")

	parsedResults := terminalW.ParseArgsCmd("l", "a", "t", "r", "h")
	// fmt.Println(parsedResults)
	coloredStrings := containerW.NewSet()
	indent := 6
	delimiter := "  "
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 4, '\t', tabwriter.AlignRight)

	var args []string

	if parsedResults == nil {
		args = []string{"./"}
		goto skipTo
	}
	args = parsedResults.Positional.ToStringSlice()

	numFileToPrint = parsedResults.GetNumArgs()
	if numFileToPrint == -1 {
		numFileToPrint = math.MaxInt32
	}

	if parsedResults.ContainsFlag("t") {
		sortType = _lsW.NewerFirst
	}

	if parsedResults.ContainsFlag("tr") || parsedResults.ContainsFlag("rt") {
		sortType = _lsW.OlderFirst
	}

	if parsedResults.ContainsFlag("a") {
		all = true
	}

	if parsedResults.ContainsFlag("l") {
		l = true
	}

	if parsedResults.ContainsFlag("h") {
		fs.PrintDefaults()
		return
	}

skipTo:
	fmt.Printf("\n")
	if len(args) == 0 {
		args = []string{"./"}
	}
	for _, rootDir := range args {
		fileMap := utilsW.LsDirGlob(rootDir)

		for d, fileSlice := range fileMap {
			files = ""
			if d != "./" && d[0] == '.' && !all {
				continue
			}
			if len(fileMap) > 1 {
				fmt.Printf("%s:\n", color.HiBlueString(d))
			}

			files += processSingleDir(d, fileSlice, l, sortType, coloredStrings)

			toPrint := stringsW.Wrap(files, w-indent*2, indent, delimiter)

			boldBlue := color.New(color.FgHiBlue, color.Bold)
			cnt := 0

			for _, line := range stringsW.SplitNoEmpty(toPrint, "\n") {
				if strings.Contains(line, "\x01") {
					line = strings.ReplaceAll(line, "\x01", "")
					fmt.Fprintln(tw, line)
					cnt++
					if cnt >= numFileToPrint {
						goto outerLoop
					}
				} else {
					fmt.Printf("\n%s", strings.Repeat(" ", indent))
					for _, word := range stringsW.SplitNoEmpty(line, delimiter) {
						word = strings.ReplaceAll(word, "\x00", " ")
						if coloredStrings.Contains(word) {
							boldBlue.Printf("%s%s", word, delimiter)
						} else {
							fmt.Printf("%s%s", word, delimiter)
						}
						cnt++
						if cnt >= numFileToPrint {
							goto outerLoop
						}
					}
					fmt.Println()
				}
			}
		outerLoop:
			tw.Flush()
			fmt.Println()
		}
	}
	printErrors()
}
