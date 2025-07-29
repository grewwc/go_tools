package main

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func calcMd5(filename string) string {
	b, _ := os.ReadFile(filename)
	return fmt.Sprintf("%x", md5.Sum(b))
}

func newFileSet(rootDir string, parser *terminalw.Parser) *cw.OrderedSetT[string] {
	s := cw.NewOrderedSetT[string]()
	files := utilsw.LsDir(rootDir, nil, nil)
	chooseExt := parser.GetFlagValueDefault("ext", "") != ""
	printMd5 := parser.ContainsFlagStrict("md5")
	for _, f := range files {
		if chooseExt {
			ext := filepath.Ext(f)
			if ext != "."+parser.GetFlagValueDefault("ext", "") {
				continue
			}
		}
		if printMd5 {
			f += fmt.Sprintf(" (%s)", calcMd5(filepath.Join(rootDir, f)))
		}
		s.Add(f)
	}
	return s
}

func main() {

	parser := terminalw.NewParser()
	parser.Bool("line", false, "if print by new line (default false)")
	parser.Bool("md5", false, "if print file md5 value (default false)")
	parser.String("ext", "", "file extension to compare (default all file types)")
	parser.Bool("h", false, "print help info")
	parser.ParseArgsCmd("-line", "-md5", "-h")
	if parser.Empty() || parser.ContainsAllFlagStrict("h") {
		parser.PrintDefaults()
		return
	}
	if len(os.Args) < 3 {
		fmt.Println("dif dir_1 dir_2")
		return
	}
	printLine := parser.ContainsFlagStrict("line")

	d1 := os.Args[1]
	d2 := os.Args[2]
	s1 := newFileSet(d1, parser)
	s2 := newFileSet(d2, parser)

	i := s1.Intersect(s2)
	s1.Subtract(i)
	s2.Subtract(i)
	sep := ", "
	if printLine {
		sep = "\n"
	}

	p1 := strings.Join(s1.ToStringSlice(), sep)
	p2 := strings.Join(s2.ToStringSlice(), sep)

	fmt.Println(p1)
	fmt.Println(strings.Repeat(".", 20))
	fmt.Println(p2)
}
