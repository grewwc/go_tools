package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/grewwc/go_tools/src/utilsW"
)

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

		err := os.Chdir(filepath.Join(utilsW.GetDirOfTheFile(), subdir))
		if err != nil {
			log.Println(err)
			continue
		}
		defer os.Chdir("../")
		filename := utilsW.LsDir(".")[0]
		executableFilename := filepath.Join(outputDir, subdir)
		if strings.ToLower(runtime.GOOS) == "windows" {
			executableFilename += ".exe"
		}
		if utilsW.IsExist(executableFilename) && utilsW.IsNewer(executableFilename, filename) {
			continue
		}

		fmt.Printf("building %q\n", filename)
		cmd := exec.Command("go", "build", "-a", "-o", outputDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			log.Println(err)
			continue
		}
	}
}
