package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

var ignoreName = cw.NewSet()
var forceRebuildName = cw.NewSet()

func init() {
	if utilsw.GetPlatform() != utilsw.WINDOWS {
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
	for fname := range parser.Positional.Iter().Iterate() {
		forceRebuildName.Add(fname.Value() + ".go")
	}

	subdirs := utilsw.LsDir(utilsw.GetDirOfTheFile(), nil, nil)
	// fmt.Println(forceRebuildName.ToSlice(), subdirs)
	outputDir := filepath.Join(utilsw.GetDirOfTheFile(), "../", "../", "bin/")
	if !utilsw.IsExist(outputDir) {
		os.MkdirAll(outputDir, os.ModePerm)
	} else if !utilsw.IsDir(outputDir) {
		log.Fatalf("cannot install, because %q is not a directory", outputDir)
	}
	for _, subdir := range subdirs {
		if !utilsw.IsDir(filepath.Join(utilsw.GetDirOfTheFile(), subdir)) || strings.Trim(subdir, " ") == "bin" {
			continue
		}

		if ignoreName.Contains(strings.TrimSpace(subdir)) {
			continue
		}

		err := os.Chdir(filepath.Join(utilsw.GetDirOfTheFile(), subdir))
		if err != nil {
			log.Println(err)
			continue
		}
		defer os.Chdir("../")
		var filename string
		filenames := utilsw.LsDir(".", nil, nil)
		// find the first go file to build as binary
		for _, name := range filenames {
			if filepath.Ext(name) != ".go" || strings.HasSuffix(name, "_test.go") {
				continue
			}
			filename = name
		}

		executableFilename := filepath.Join(outputDir, utilsw.TrimFileExt(filename))
		if utilsw.GetPlatform() == utilsw.WINDOWS {
			executableFilename += ".exe"
		}
		// fmt.Println("what", filename, forceRebuildName)
		// fmt.Println(filename, forceRebuildName.Contains(filename))
		if (!all && !force && !forceRebuildName.Contains(filename)) &&
			(utilsw.IsExist(executableFilename) && utilsw.IsNewer(executableFilename, filename)) {
			continue
		}

		fmt.Printf("building %q\n", filename)
		cmd := fmt.Sprintf(`go build -ldflags="-s -w" -o %s`, filepath.Join(outputDir, filepath.Base(executableFilename)))
		if _, err := utilsw.RunCmd(cmd, nil); err != nil {
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
