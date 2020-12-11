package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

func install() {
	flag.Bool("h", false, "print help information")
	flag.Bool("f", false, "force rebuild (shortcut form)")
	flag.Bool("force", false, "force rebuilds")
	parsed := terminalW.ParseArgsCmd("h", "-f", "force")
	if parsed != nil && parsed.ContainsFlag("h") {
		flag.PrintDefaults()
		return
	}
	filename := filepath.Join(utilsW.GetDirOfTheFile(), "src", "executable", "build.go")
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
