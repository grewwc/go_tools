package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	defaultN = 5
)

const (
	logHistoryCmd = `git log --oneline --format="%h %an %ad %s" --date=short`
	branchCmd     = `git for-each-ref --sort=-committerdate --format="%(refname:short) %(committerdate:short) %(subject)" refs/heads/ `
)

type ILineHandler interface {
	handleLine(string)
}

type logHandler struct {
	p *regexp.Regexp
}

func (h *logHandler) handleLine(line string) {
	if matched := h.p.MatchString(line); !matched {
		fmt.Println(line)
	}
}

type branchHandler struct {
}

func (h *branchHandler) handleLine(line string) {
	if line == "" {
		return
	}
	parts := strw.SplitNoEmpty(line, " ")
	if len(parts) < 3 {
		fmt.Println(line)
		return
	}
	branchName := color.BlueString(parts[0])
	modifyTime := color.YellowString(parts[1])
	subject := strings.Join(parts[2:], " ")
	fmt.Printf("%s (%s) %s\n", branchName, modifyTime, subject)
}

func getN(parser *terminalw.Parser) int {
	if parser.Empty() {
		return -1
	}
	n := parser.GetNumArgs()
	if n != -1 {
		return n
	}
	n, err := strconv.Atoi(parser.GetFlagValueDefault("n", "-1"))
	if err != nil {
		panic(err)
	}
	if n != -1 {
		return n
	}
	if parser.ContainsFlagStrict("a") {
		return math.MaxInt
	}
	return n
}

func getCmd(parser *terminalw.Parser) string {
	if parser.ContainsFlagStrict("b") {
		return branchCmd
	}
	return logHistoryCmd
}

func getHandler(parser *terminalw.Parser) ILineHandler {
	if parser.ContainsAllFlagStrict("b") {
		return &branchHandler{}
	}
	return &logHandler{
		p: regexp.MustCompile(`\w+\s.*\s\d{4}-\d{2}-\d{2}\sMerge.*`),
	}
}

func handleLine(h ILineHandler, line string) {
	h.handleLine(line)
}

func main() {
	parser := terminalw.NewParser()
	parser.Int("n", defaultN, "num of histories to print")
	parser.Bool("a", false, "print all histories")
	parser.Bool("h", false, "print help info")
	parser.Bool("b", false, "print branch")
	parser.ParseArgsCmd("h", "a", "b")
	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}
	n := getN(parser)
	if n == -1 {
		n = defaultN
	}
	cmd := getCmd(parser)
	handler := getHandler(parser)
	res, err := utilsw.RunCmd(cmd, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
	cnt := 0
	for line := range strw.SplitByToken(strings.NewReader(res), "\n", false) {
		if cnt >= n {
			break
		}
		handleLine(handler, line)
		cnt++
	}
}
