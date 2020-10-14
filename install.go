package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/grewwc/go_tools/src/utilsW"
)

func install() {
	filename := filepath.Join(utilsW.GetDirOfTheFile(), "src", "executable", "build.go")
	cmd := exec.Command("go", "run", filename)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	install()
}
