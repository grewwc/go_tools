package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw"
)

var (
	force   bool
	verbose bool
	newline bool
)

var (
	e  *cw.Set
	ne *cw.Set
)

type iTask func(name string) error

func truncFile(name string) error {
	return os.Truncate(name, 0)
}

func removeNewLine(name string) error {
	lines := make([]string, 0)
	f, err := os.Open(name)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	for {
		b, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}
		if len(b) > 0 {
			line := typesw.BytesToStr(bytes.Trim(b, "\n"))
			if typesw.BytesToStr(bytes.TrimSpace(b)) != "" {
				lines = append(lines, line)
			}
		}
		if err == io.EOF {
			break
		}
	}
	utilsw.WriteToFile(name, typesw.StrToBytes(strings.Join(lines, "\n")))
	return nil
}

func needTruncate(ext string) bool {
	if e == nil && ne == nil {
		return true
	}
	if e == nil {
		return !ne.Contains(ext)
	}
	return e.Contains(ext)
}

func truncateDirOrFile(name string, task iTask) error {
	if !force && !newline && !utilsw.PromptYesOrNo(color.RedString("Are you sure to truncate all files in %q (y/n) ", name)) {
		fmt.Println("Aborting...")
		return nil
	}
	if utilsw.IsDir(name) {
		if err := filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}
			if verbose {
				fmt.Printf(" -- truncating %q\n", path)
			}
			ext := strings.TrimLeft(filepath.Ext(path), ".")
			// fmt.Println(needTruncate(ext), ext, e, ne)
			if needTruncate(ext) {
				if err := os.Truncate(path, 0); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}
	ext := strings.TrimLeft(filepath.Ext(name), ".")
	// fmt.Println(needTruncate(ext), ext, e, ne)
	if needTruncate(ext) {
		return task(name)
	}
	return nil
}

func printHelp() {
	fmt.Println("truncate dir/filenaame")
}

func getStringSlice(s string) []string {
	s = strings.ReplaceAll(s, ",", " ")
	return strw.SplitNoEmpty(s, " ")
}

func main() {
	parser := terminalw.NewParser()
	parser.Bool("f", false, "force")
	parser.Bool("v", false, "verbose")
	parser.Bool("h", false, "print help info")
	parser.Bool("newline", false, "only remove newline")
	parser.String("include", "", "only trucnate files with the extension, e.g.: -include \".log, .txt\"")
	parser.String("exclude", "", "only trucnate files without the extension, e.g.: -exclude \".log, .txt\"")
	parser.ParseArgsCmd("v", "h", "f", "newline")
	var root string
	var err error
	if parser.Empty() || parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		printHelp()
		return
	}
	force = parser.ContainsFlagStrict("f")
	verbose = parser.ContainsFlagStrict("v")
	includeExt := parser.GetFlagValueDefault("include", "")
	excludeExt := parser.GetFlagValueDefault("exclude", "")
	newline = parser.ContainsFlagStrict("newline")
	if includeExt != "" {
		e = cw.NewSet()
		for _, ext := range getStringSlice(includeExt) {
			e.Add(strings.TrimLeft(ext, "."))
		}
	}
	if excludeExt != "" {
		ne = cw.NewSet()
		for _, ext := range getStringSlice(excludeExt) {
			ne.Add(strings.TrimLeft(ext, "."))
		}
	}
	pos := parser.Positional.ToStringSlice()
	if len(pos) > 1 {
		fmt.Println(color.RedString("atmost 1 arg"))
		return
	}
	root = pos[0]
	task := truncFile
	if newline {
		task = removeNewLine
	}
	if err = truncateDirOrFile(root, task); err != nil {
		panic(err)
	}
}
