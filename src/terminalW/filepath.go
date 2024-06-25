package terminalW

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"unicode"
)

func buildRegex(filename string) string {
	res := ""
	for _, c := range filename {
		next := string(c)
		if unicode.IsLetter(c) {
			res += fmt.Sprintf("[%s%s]", strings.ToLower(next), strings.ToUpper(next))
		} else {
			res += next
		}
	}
	return res
}

func Glob(pattern, rootPath string) ([]string, error) {
	pattern = path.Join(rootPath, pattern)
	return filepath.Glob(pattern)
}

func GlobCaseInsensitive(pattern, rootPath string) ([]string, error) {
	pattern = buildRegex(pattern)
	pattern = path.Join(rootPath, pattern)
	// fmt.Println(pattern)
	return filepath.Glob(pattern)
}
