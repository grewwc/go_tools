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
var ignores = containerW.NewSet()

var errMsgs = containerW.NewQueue()

func init() {
	windowsW.EnableVirtualTerminal()
	var info windows.ConsoleScreenBufferInfo
	stdout := windows.Handle(os.Stdout.Fd())

	windows.GetConsoleScreenBufferInfo(stdout, &info)
	w = int(info.Size.X)
}

func formatFileStat(filename string, realSize bool) string {
	stat, err := os.Lstat(filename)
	if err != nil {
		errMsgs.Enqueue(fmt.Sprintf("error getting stat of file: %q\n", filename))
		return ""
		// os.Exit(1)
	}
	modTime := stat.ModTime()
	modTimeStr := fmt.Sprintf("   %04d/%02d/%02d  %02d:%02d", modTime.Year(), int(modTime.Month()), modTime.Day(), modTime.Hour(), modTime.Minute())
	var sizeStr string
	if !utilsW.IsDir(filename) {
		sizeStr = stringsW.FormatInt64(stat.Size())
	} else {
		var dirSize int64
		var err error
		if realSize {
			dirSize, err = utilsW.GetDirSize(filename)
		} else {
			dirSize = stat.Size()
		}
		if err != nil {
			errMsgs.Enqueue(fmt.Sprintf("error getting size of directory: %q\n", filename))
			return ""
			// log.Printf("error getting size of directory: %q\n", filename)
			// os.Exit(1)
		}
		sizeStr = stringsW.FormatInt64(dirSize)
	}
	if utilsW.IsDir(filename) {
		filename = color.HiCyanString(filename + "/")
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

func processSingleDir(rootDir string, fileSlice []string, long bool, du bool, sortType int,
	coloredStrings *containerW.Set) string {

	// if sortType != _lsW.Unsort
	// sort the fileSlice
	if sortType != _lsW.Unsort {
		fileSlice = _lsW.SortByModifiedDate(fileSlice, sortType)
	}
	// fmt.Println("here", fileSlice, sortType)
	files := ""
	for _, file := range fileSlice {
		ext := filepath.Ext(file)
		if ignores.Contains(ext) {
			continue
		}
		file = filepath.Join(rootDir, file)
		file = filepath.ToSlash(file)
		if !all && filepath.Base(file)[0] == '.' {
			continue
		}
		if long {
			line := formatFileStat(file, du)
			if rootDir[len(rootDir)-1] != '/' {
				rootDir += "/"
			}
			line = strings.Replace(line, rootDir, "", 1)
			// fmt.Println("long", file, line)
			if line != "" {
				files += line + "\x01\n"
			}
			continue
		}
		if utilsW.IsDir(file) {
			file += "/"
			if rootDir[len(rootDir)-1] != '/' {
				rootDir += "/"
			}
			coloredStrings.Add(stringsW.StripPrefix(file, rootDir))
		}
		if strings.Contains(file, " ") {

			if rootDir[len(rootDir)-1] != '/' {
				rootDir += "/"
			}
			file = stringsW.StripPrefix(file, rootDir)
			fileWithQuote := fmt.Sprintf("\"%s\"", file)
			if utilsW.IsDir(file) {
				coloredStrings.Add(fileWithQuote)
			}

			// later on, string will be seperated by space, we
			// have to replace space with \x00
			file = strings.ReplaceAll(fileWithQuote, " ", "\x00")
		}
		if rootDir[len(rootDir)-1] != '/' {
			rootDir += "/"
		}
		file = stringsW.StripPrefix(file, rootDir)
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
	var du bool
	var moreIgnores string

	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.Bool("l", false, "show more detailed information")
	fs.Bool("a", false, "list hidden file")
	fs.Bool("t", false, "sort files by last modified date")
	fs.Bool("rt", false, "sort files by earlist modified date")
	fs.Bool("h", false, "print help information")
	fs.Bool("du", false, "if set, calculate size of all subdirs/subfiles")
	fs.String("nt", "", "types/extensions that will not be listed. e.g.: -nt \"py png, jpg\"")
	parsedResults := terminalW.ParseArgsCmd("l", "a", "t", "r", "du")
	// parsedResults := terminalW.ParseArgsCmd()

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

	moreIgnores, _ = parsedResults.GetFlagVal("nt")
	moreIgnores = strings.ReplaceAll(moreIgnores, ",", " ")
	for _, moreIgnore := range stringsW.SplitNoEmpty(moreIgnores, " ") {
		if moreIgnore[0] != '.' {
			moreIgnore = "." + moreIgnore
		}
		ignores.Add(moreIgnore)
	}

	if parsedResults.ContainsFlag("t") {
		if !parsedResults.ContainsFlag("r") {
			sortType = _lsW.NewerFirst
		} else {
			sortType = _lsW.OlderFirst
		}
	}

	if parsedResults.ContainsFlag("a") {
		all = true
	}

	if parsedResults.ContainsFlag("l") {
		l = true
	}

	if parsedResults.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}

	if parsedResults.ContainsFlag("du") {
		du = true
	}

skipTo:
	fmt.Printf("\n")
	if len(args) == 0 {
		args = []string{"./"}
	}

	// fmt.Println(args)
	for _, rootDir := range args {
		if len(args) > 1 {
			fmt.Printf("%s:\n", color.HiCyanString(rootDir))
		}
		fileMap := utilsW.LsDirGlob(rootDir)
		// fmt.Println("filemap: ", fileMap)
		for d, fileSlice := range fileMap {
			files = ""
			if !strings.HasPrefix(d, "./") &&
				!strings.HasPrefix(d, "../") &&
				d[0] == '.' && !all {
				continue
			}
			if len(fileMap) > 1 {
				fmt.Printf("%s:\n", color.HiCyanString(d))
			}
			files += processSingleDir(d, fileSlice, l, du, sortType, coloredStrings)
			// fmt.Println("file: ===>", files)
			var toPrint string = files
			if !l {
				toPrint = stringsW.Wrap(files, w-indent+2, indent, delimiter)
			}
			boldCyan := color.New(color.FgHiCyan, color.Bold)
			cnt := 0

			for _, line := range stringsW.SplitNoEmpty(toPrint, "\n") {
				if strings.Contains(line, "\x01") { // \x01 means ls -l
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
							boldCyan.Printf("%s%s", word, delimiter)
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
