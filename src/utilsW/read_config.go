package utilsW

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

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

func GetAllConfig() (m map[string]string) {
	f, err := os.Open(fname)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	m = make(map[string]string)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		res := stringsW.SplitNoEmptyKeepQuote(line, '=')
		// fmt.Println(res)
		key, val := res[0], res[1]
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		m[key] = val
	}
	return
}
