package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

func main() {
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.String("ex", "", "exclude file/directory")
	fs.String("exclude", "", "exclude file/directory")
	fs.Bool("v", false, "verbose")

	parsedResults := terminalW.ParseArgsCmd("v")
	if parsedResults == nil {
		fs.PrintDefaults()
		return
	}
	exclude, err := parsedResults.GetFlagVal("ex")
	if err != nil || exclude == "" {
		exclude, _ = parsedResults.GetFlagVal("exclude")
	}
	exclude = utilsW.Abs(exclude)

	excludes, err := filepath.Glob(exclude)
	if err != nil {
		log.Println(err)
		return
	}

	verbose := parsedResults.ContainsFlag("v")
	excludeSet := containerW.NewSet()
	for _, ex := range excludes {

		filepath.Walk(ex, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}
			excludeSet.Add(path)
			return nil
		})
	}

	args := parsedResults.Positional.ToStringSlice()
	srcNames := []string{}
	var srcName string
	outName := args[0]
	if len(args) > 2 {
		srcNames = args[1:]
	} else {
		srcName = utilsW.Abs(args[1])
	}

	if srcName != "" {
		srcNames, err = filepath.Glob(srcName)
	}

	if err != nil {
		log.Fatalln(err)
	}

	allFiles := []string{}
	for _, srcName := range srcNames {
		filepath.Walk(srcName, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			path = utilsW.Abs(path)
			if !excludeSet.Contains(path) {
				allFiles = append(allFiles, path)
				if verbose {
					fmt.Println(path)
				}
			} else if verbose {
				fmt.Println("exclude: ", path)
			}
			return nil
		})
	}
	if len(allFiles) == 0 {
		fmt.Printf("%q don't contain any files\n", srcName)
		return
	}

	if err = utilsW.TarGz(outName, allFiles); err != nil {
		log.Fatalln(err)
	}
	fmt.Println()
}
