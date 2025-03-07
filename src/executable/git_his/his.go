package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	defaultN = 5
)

func getN(parsed *terminalW.ParsedResults) int {
	if parsed.Empty() {
		return -1
	}
	n := parsed.GetNumArgs()
	if n != -1 {
		return n
	}
	n, err := strconv.Atoi(parsed.GetFlagValueDefault("n", "-1"))
	if err != nil {
		panic(err)
	}
	if n != -1 {
		return n
	}
	if parsed.ContainsFlagStrict("a") {
		return math.MaxInt
	}
	return n
}

func main() {
	flag.Int("n", defaultN, "num of histories to print")
	flag.Bool("a", false, "print all histories")
	flag.Bool("h", false, "print help info")
	parsed := terminalW.ParseArgsCmd("h", "a")
	if parsed.ContainsFlagStrict("h") {
		flag.PrintDefaults()
		return
	}
	n := getN(parsed)
	if n == -1 {
		n = defaultN
	}
	cmd := `git log --oneline --format="%h %an %ad %s" --date=short`
	pattern := `\w+\s.*\s\d{4}-\d{2}-\d{2}\sMerge.*`
	res, err := utilsW.RunCmd(cmd, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return
	}
	p := regexp.MustCompile(pattern)
	cnt := 0
	for history := range stringsW.SplitByToken(strings.NewReader(res), "\n", false) {
		if cnt >= n {
			break
		}
		if matched := p.MatchString(history); !matched {
			fmt.Println(history)
			cnt++
		}
	}
}
