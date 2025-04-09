package terminalw

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/grewwc/go_tools/src/conw"
	"github.com/grewwc/go_tools/src/strw"
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
			res.WriteString(fmt.Sprintf(" %s \"%s\" ", k, v))
		} else {
			res.WriteString(fmt.Sprintf(" %s %s ", k, v))
		}
	}
	return res.String()
}

// support "," and " " split
func FormatFileExtensions(extensions string) *conw.Set {
	extensions = strings.ReplaceAll(extensions, ",", " ")
	bySpace := strw.SplitNoEmpty(extensions, " ")

	var res = conw.NewSet()
	for _, ext := range bySpace {
		if !strings.HasPrefix(ext, ".") {
			res.Add("." + ext)
		} else {
			res.Add(ext)
		}
	}
	return res
}
