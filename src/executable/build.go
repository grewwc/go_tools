package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

var ignoreName = containerW.NewSet()
var forceRebuildName = containerW.NewSet()

func init() {
	if utilsW.GetPlatform() != utilsW.WINDOWS {
		// add Folder name, NOT filename
		ignoreName.AddAll("cat", "head", "rm", "stat", "tail", "touch", "open", "ls")
	}
}

func main() {
	parser := terminalW.NewParser()
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

	subdirs := utilsW.LsDir(utilsW.GetDirOfTheFile(), nil, nil)
	// fmt.Println(forceRebuildName.ToSlice(), subdirs)
	outputDir := filepath.Join(utilsW.GetDirOfTheFile(), "../", "../", "bin/")
	if !utilsW.IsExist(outputDir) {
		os.MkdirAll(outputDir, os.ModePerm)
	} else if !utilsW.IsDir(outputDir) {
		log.Fatalf("cannot install, because %q is not a directory", outputDir)
	}
	for _, subdir := range subdirs {
		if !utilsW.IsDir(filepath.Join(utilsW.GetDirOfTheFile(), subdir)) || strings.Trim(subdir, " ") == "bin" {
			continue
		}

		if ignoreName.Contains(strings.TrimSpace(subdir)) {
			continue
		}

		err := os.Chdir(filepath.Join(utilsW.GetDirOfTheFile(), subdir))
		if err != nil {
			log.Println(err)
			continue
		}
		defer os.Chdir("../")
		var filename string
		filenames := utilsW.LsDir(".", nil, nil)
		// find the first go file to build as binary
		for _, name := range filenames {
			if filepath.Ext(name) != ".go" {
				continue
			}
			filename = name
		}

		executableFilename := filepath.Join(outputDir, utilsW.TrimFileExt(filename))
		if utilsW.GetPlatform() == utilsW.WINDOWS {
			executableFilename += ".exe"
		}
		// fmt.Println("what", filename, forceRebuildName)
		if (!all && !force && !forceRebuildName.Contains(filename)) &&
			(utilsW.IsExist(executableFilename) && utilsW.IsNewer(executableFilename, filename)) {
			continue
		}

		fmt.Printf("building %q\n", filename)
		cmd := fmt.Sprintf(`go build -a -ldflags="-s -w" -o %s`, filepath.Join(outputDir, filepath.Base(executableFilename)))
		if _, err := utilsW.RunCmd(cmd, nil); err != nil {
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
