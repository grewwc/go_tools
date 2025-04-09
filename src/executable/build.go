package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/conw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilw"
)

var ignoreName = conw.NewSet()
var forceRebuildName = conw.NewSet()

func init() {
	if utilw.GetPlatform() != utilw.WINDOWS {
		// add Folder name, NOT filename
		ignoreName.AddAll("cat", "head", "rm", "stat", "tail", "touch", "open", "ls")
	}
}

func main() {
	parser := terminalw.NewParser()
	parser.Bool("h", false, "print help information")
	parser.Bool("f", false, "force rebuild (shortcut form)")
	parser.Bool("a", false, "force rebuild all")
	parser.Bool("force", false, "force rebuilds")
	parser.ParseArgsCmd("f", "force", "a", "h")
	var force bool = parser.ContainsFlag("f") || parser.ContainsFlag("force")
	var all bool = parser.ContainsFlagStrict("a")
	for fname := range parser.Positional.Iterate() {
		fnameStr := fname.(string)
		forceRebuildName.Add(fnameStr + ".go")
	}

	subdirs := utilw.LsDir(utilw.GetDirOfTheFile(), nil, nil)
	// fmt.Println(forceRebuildName.ToSlice(), subdirs)
	outputDir := filepath.Join(utilw.GetDirOfTheFile(), "../", "../", "bin/")
	if !utilw.IsExist(outputDir) {
		os.MkdirAll(outputDir, os.ModePerm)
	} else if !utilw.IsDir(outputDir) {
		log.Fatalf("cannot install, because %q is not a directory", outputDir)
	}
	for _, subdir := range subdirs {
		if !utilw.IsDir(filepath.Join(utilw.GetDirOfTheFile(), subdir)) || strings.Trim(subdir, " ") == "bin" {
			continue
		}

		if ignoreName.Contains(strings.TrimSpace(subdir)) {
			continue
		}

		err := os.Chdir(filepath.Join(utilw.GetDirOfTheFile(), subdir))
		if err != nil {
			log.Println(err)
			continue
		}
		defer os.Chdir("../")
		var filename string
		filenames := utilw.LsDir(".", nil, nil)
		// find the first go file to build as binary
		for _, name := range filenames {
			if filepath.Ext(name) != ".go" {
				continue
			}
			filename = name
		}

		executableFilename := filepath.Join(outputDir, utilw.TrimFileExt(filename))
		if utilw.GetPlatform() == utilw.WINDOWS {
			executableFilename += ".exe"
		}
		// fmt.Println("what", filename, forceRebuildName)
		if (!all && !force && !forceRebuildName.Contains(filename)) &&
			(utilw.IsExist(executableFilename) && utilw.IsNewer(executableFilename, filename)) {
			continue
		}

		fmt.Printf("building %q\n", filename)
		cmd := fmt.Sprintf(`go build -a -ldflags="-s -w" -o %s`, filepath.Join(outputDir, filepath.Base(executableFilename)))
		if _, err := utilw.RunCmd(cmd, nil); err != nil {
			panic(err)
		}

		// cmd := exec.Command("go", "build", "-a", "-o", filepath.Join(outputDir, filepath.Base(executableFilename)))
		// cmd.Stdout = os.Stdout
		// cmd.Stderr = os.Stderr
		// err = cmd.Run()
		// if err != nil {
		// 	log.Println(err)
		// 	continue
		// }
	}
}
