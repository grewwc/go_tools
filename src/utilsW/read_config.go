package utilsW

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
)

var (
	fname = os.Getenv("HOME")
)

func init() {
	if strings.TrimSpace(fname) == "" {
		log.Fatalln("HOME not set")
	}
	fname = filepath.Join(fname, ".configW")
}

func GetAllConfig() (m *containerW.OrderedMap) {
	f, err := os.Open(fname)
	if err != nil {
		fmt.Printf("%s not found, ignored...", color.RedString(fname))
	}
	defer f.Close()
	m = containerW.NewOrderedMap()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		trimed := strings.TrimSpace(line)
		line = stringsW.TrimAfter(line, "#")
		line = stringsW.TrimAfter(line, "//")
		// comment
		if stringsW.AnyHasPrefix(trimed, "#", "//") {
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}

		res := stringsW.SplitNoEmptyKeepQuote(line, '=')
		var key, val string
		// fmt.Println(res)
		key = res[0]
		key = strings.TrimSpace(key)
		if len(res) > 1 {
			val = res[1]
			val = strings.TrimSpace(val)
		}
		m.Put(key, val)
	}
	return
}
