package terminalW

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"golang.org/x/sys/windows"
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
func FormatFileExtensions(extensions string) *containerW.Set {
	extensions = strings.ReplaceAll(extensions, ",", " ")
	bySpace := stringsW.SplitNoEmpty(extensions, " ")

	var res = containerW.NewSet()
	for _, ext := range bySpace {
		if !strings.HasPrefix(ext, ".") {
			res.Add("." + ext)
		} else {
			res.Add(ext)
		}
	}
	return res
}

func EnableVirtualTerminal() {
	stdout := windows.Handle(os.Stdout.Fd())
	var originalMode uint32

	windows.GetConsoleMode(stdout, &originalMode)
	windows.SetConsoleMode(stdout, originalMode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)

}
