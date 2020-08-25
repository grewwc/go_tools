package terminalW

import (
	"bytes"
	"fmt"
	"go_tools/src/stringsW"
	"strings"
)

func AddQuote(slice []string) []string {
	var res []string
	for _, s := range slice {
		res = append(res, fmt.Sprintf("%q", s))
	}
	return res
}

func MapToString(m map[string]string) string {
	var res bytes.Buffer
	for k, v := range m {
		if strings.Contains(strings.TrimSpace(v), " ") {
			res.WriteString(fmt.Sprintf(" -%s \"%s\" ", k, v))
		} else {
			res.WriteString(fmt.Sprintf(" -%s %s ", k, v))
		}
	}
	return res.String()
}

// support "," and " " split
func FormatFileExtensions(extensions string) string {
	extensions = strings.ReplaceAll(extensions, ",", " ")
	bySpace := stringsW.SplitNoEmpty(extensions, " ")

	var res string
	for _, ext := range bySpace {
		if !strings.HasPrefix(ext, ".") {
			res += ("." + ext)
		} else {
			res += ext
		}
	}
	return res
}
