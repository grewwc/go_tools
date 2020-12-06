package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"github.com/grewwc/go_tools/src/windowsW"
	"golang.org/x/sys/windows"
)

var w int
var all *bool

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

func processSingleDir(rootDir string, fileSlice []string, long bool,
	coloredStrings *containerW.Set) string {

	files := ""
	for _, file := range fileSlice {
		file = filepath.Join(rootDir, file)
		file = filepath.ToSlash(file)
		if !*all && filepath.Base(file)[0] == '.' {
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
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	l := fs.Bool("l", false, "show more information")
	all = fs.Bool("a", false, "list hidden file")
	lall := fs.Bool("la", false, "shortcut for -l -a")
	alll := fs.Bool("al", false, "shortcut for -l -a")

	parsedResults := terminalW.ParseArgsCmd("l", "a", "al", "la")
	coloredStrings := containerW.NewSet()
	indent := 6
	delimiter := "  "
	tw := tabwriter.NewWriter(os.Stdout, 0, 8, 4, '\t', tabwriter.AlignRight)
	var optionalStr string
	var optional map[string]string
	var args []string

	if parsedResults == nil {
		args = []string{"./"}
		goto skip
	}
	optional, args = parsedResults.Optional, parsedResults.Positional.ToStringSlice()
	// fmt.Println("optional", optional)
	// fmt.Println("positional", args)
	optionalStr = terminalW.MapToString(optional)
	fs.Parse(stringsW.SplitNoEmptyKeepQuote(optionalStr, ' '))

	if *lall || *alll {
		*l = true
		*all = true
	}

skip:
	fmt.Printf("\n")
	for _, rootDir := range args {
		fileMap := utilsW.LsDirGlob(rootDir)
		for d, fileSlice := range fileMap {
			files = ""
			if d != "./" && d[0] == '.' && !*all {
				continue
			}
			if len(fileMap) > 1 {
				fmt.Printf("%s:\n", color.HiBlueString(d))
			}
			files += processSingleDir(d, fileSlice, *l, coloredStrings)

			toPrint := stringsW.Wrap(files, w-indent*2, indent, delimiter)

			boldBlue := color.New(color.FgHiBlue, color.Bold)
			for _, line := range stringsW.SplitNoEmpty(toPrint, "\n") {
				if strings.Contains(line, "\x01") {
					line = strings.ReplaceAll(line, "\x01", "")
					fmt.Fprintln(tw, line)
				} else {
					fmt.Printf("\n%s", strings.Repeat(" ", indent))
					for _, word := range stringsW.SplitNoEmpty(line, delimiter) {
						word = strings.ReplaceAll(word, "\x00", " ")
						if coloredStrings.Contains(word) {
							boldBlue.Printf("%s%s", word, delimiter)
						} else {
							fmt.Printf("%s%s", word, delimiter)
						}
					}
					fmt.Println()
				}
			}
			tw.Flush()
			fmt.Println()
		}
	}
	printErrors()
}
