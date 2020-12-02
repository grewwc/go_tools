package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/utilsW"
)

var ignoreName = containerW.NewSet()

func init() {
	if utilsW.GetPlatform() != utilsW.WINDOWS {
		ignoreName.AddAll("ls", "cat")
	}
}

func main() {
	subdirs := utilsW.LsDir(utilsW.GetDirOfTheFile())
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
		filename := utilsW.LsDir(".")[0]
		executableFilename := filepath.Join(outputDir, utilsW.TrimFileExt(filename))
		if utilsW.GetPlatform() == utilsW.WINDOWS {
			executableFilename += ".exe"
		}
		if utilsW.IsExist(executableFilename) && utilsW.IsNewer(executableFilename, filename) {
			continue
		}

		fmt.Printf("building %q\n", filename)
		cmd := exec.Command("go", "build", "-a", "-o", filepath.Join(outputDir, filepath.Base(executableFilename)))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
