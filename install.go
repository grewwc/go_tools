package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilw"
)

func install() {
	parser := terminalw.NewParser()
	parser.Bool("h", false, "print help information")
	parser.Bool("f", false, "force rebuild (shortcut form)")
	parser.Bool("a", false, "force rebuild all")
	parser.Bool("force", false, "force rebuilds")
	parser.ParseArgsCmd("h", "-f", "force", "a")
	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}
	filename := filepath.Join(utilw.GetDirOfTheFile(), "src", "executable", "build.go")
	var args = []string{"run", filename}
	for _, additional := range os.Args[1:] {
		args = append(args, additional)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalln(err)
	}
}

func main() {
	install()
}
