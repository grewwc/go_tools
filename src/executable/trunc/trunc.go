package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

var (
	force   bool
	verbose bool
)

var (
	e  *containerW.Set
	ne *containerW.Set
)

func needTruncate(ext string) bool {
	if e == nil && ne == nil {
		return true
	}
	if e == nil {
		return !ne.Contains(ext)
	}
	return e.Contains(ext)
}

func truncateDirOrFile(name string) error {
	if !force && !utilsW.PromptYesOrNo(color.RedString("Are you sure to truncate all files in %q (y/n) ", name)) {
		fmt.Println("Aborting...")
		return nil
	}
	if utilsW.IsDir(name) {
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
		return os.Truncate(name, 0)
	}
	return nil
}

func printHelp() {
	fmt.Println("truncate dir/filenaame")
}

func getStringSlice(s string) []string {
	s = strings.ReplaceAll(s, ",", " ")
	return stringsW.SplitNoEmpty(s, " ")
}

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("f", false, "force")
	fs.Bool("v", false, "verbose")
	fs.String("e", "", "only trucnate files with the extension, e.g.: -e \".log, .txt\"")
	fs.String("ne", "", "only trucnate files with the extension, e.g.: -e \".log, .txt\"")

	parsed := terminalW.ParseArgsCmd("v", "h", "f")
	var root string
	var err error
	if parsed == nil {
		fs.PrintDefaults()
		printHelp()
		return
	}
	force = parsed.ContainsFlagStrict("f")
	verbose = parsed.ContainsFlagStrict("v")
	includeExt := parsed.GetFlagValueDefault("e", "")
	excludeExt := parsed.GetFlagValueDefault("ne", "")
	if includeExt != "" {
		e = containerW.NewSet()
		for _, ext := range getStringSlice(includeExt) {
			e.Add(strings.TrimLeft(ext, "."))
		}
	}
	if excludeExt != "" {
		ne = containerW.NewSet()
		for _, ext := range getStringSlice(excludeExt) {
			ne.Add(strings.TrimLeft(ext, "."))
		}
	}
	if parsed.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		printHelp()
		return
	}
	pos := parsed.Positional.ToStringSlice()
	if len(pos) > 1 {
		fmt.Println(color.RedString("atmost 1 arg"))
		return
	}
	root = pos[0]

	if err = truncateDirOrFile(root); err != nil {
		panic(err)
	}
}
