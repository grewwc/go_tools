package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

func calcMd5(filename string) string {
	b, _ := os.ReadFile(filename)
	return fmt.Sprintf("%x", md5.Sum(b))
}

func newFileSet(rootDir string, parsedResults *terminalW.ParsedResults) *containerW.OrderedSet {
	s := containerW.NewOrderedSet()
	files := utilsW.LsDir(rootDir, nil)
	chooseExt := parsedResults.GetFlagValueDefault("ext", "") != ""
	printMd5 := parsedResults.ContainsFlagStrict("md5")
	for _, f := range files {
		if chooseExt {
			ext := filepath.Ext(f)
			if ext != "."+parsedResults.GetFlagValueDefault("ext", "") {
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

	flag.Bool("line", false, "if print by new line (default false)")
	flag.Bool("md5", false, "if print file md5 value (default false)")
	flag.String("ext", "", "file extension to compare (default all file types)")
	flag.Bool("h", false, "print help info")
	parsedResults := terminalW.ParseArgsCmd("-line", "-md5", "-h")
	if parsedResults == nil || parsedResults.ContainsAllFlagStrict("h") {
		flag.PrintDefaults()
		return
	}
	if len(os.Args) < 3 {
		fmt.Println("dif dir_1 dir_2")
		return
	}
	printLine := parsedResults.ContainsFlagStrict("line")

	d1 := os.Args[1]
	d2 := os.Args[2]
	s1 := newFileSet(d1, parsedResults)
	s2 := newFileSet(d2, parsedResults)

	i := s1.Intersect(*s2)
	s1.Subtract(*i)
	s2.Subtract(*i)
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
