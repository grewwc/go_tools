package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func getPythonCommand() string {
	conf := utilsw.GetAllConfig()
	if conf == nil {
		return ""
	}
	cmd := conf.GetOrDefault("utils.oo.cmd", nil)
	if cmd == nil {
		log.Println(`"utils.oo.cmd" not found in ~/.configW, so only support text copy/paste`)

		return ""
	}
	return cmd.(string)
}

func exePythonScript(cmd string) {
	cmd += " " + strings.Join(os.Args[1:], " ")
	_, err := utilsw.RunCmdWithTimeout(cmd, time.Second*60)
	if err != nil {
		log.Fatalln(err)
		return
	}
}

func fileToClipboard(parser *terminalw.Parser) {
	pos := parser.GetPositionalArgs(true)
	if len(pos) != 1 {
		log.Fatalln("should input the filename")
	}
	filename := pos[0]
	utilsw.WriteClipboardText(utilsw.ReadString(filename))
}

func clipboardToFile(parser *terminalw.Parser) {
	pos := parser.GetPositionalArgs(true)
	if len(pos) >= 1 {
		filename := pos[0]
		utilsw.WriteToFile(filename, []byte(utilsw.ReadClipboardText()))
	} else {
		io.Copy(os.Stdout, strings.NewReader(utilsw.ReadClipboardText()))
	}
}

func inputToClipboard() {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, os.Stdin); err != nil {
		log.Fatalln(err)
	}
	utilsw.WriteClipboardText(buf.String())
}

func main() {
	parser := terminalw.NewParser()
	parser.Bool("i", false, "copy from stdin to clipboard")
	parser.Bool("c", false, "copy from file to clipboard")
	parser.Bool("p", false, "paste from clipboard to file")
	parser.Bool("b", false, "binary file")
	parser.Bool("h", false, "print help info")

	parser.ParseArgsCmd()

	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}

	// use python script
	cmd := getPythonCommand()
	if cmd != "" && parser.ContainsFlag("b") {
		exePythonScript(cmd)
		return
	}

	// file to clipboard
	parser.On(func(p *terminalw.Parser) bool {
		return p.ContainsFlagStrict("c")
	}).Do(func() {
		fileToClipboard(parser)
	})

	// clipboard to file
	parser.On(func(p *terminalw.Parser) bool {
		return p.ContainsFlagStrict("p")
	}).Do(func() {
		clipboardToFile(parser)
	})

	// stdin to clipboard
	parser.On(func(p *terminalw.Parser) bool {
		return p.Empty() || p.ContainsFlagStrict("i")
	}).Do(inputToClipboard)

	parser.Execute()
}
