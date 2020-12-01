package stringsW

import (
	"fmt"
	"strings"
)

func Wrap(s string, width int, indent int) (string, error) {
	words := SplitNoEmpty(s, " ")
	lines := make([]string, 0, 8)
	curLine := make([]string, 0, 8)
	cursor := indent
	indentString := strings.Repeat(" ", indent)

	for _, word := range words {
		word = indentString + strings.TrimSpace(word)
		if cursor+len(word) > width {
			lines = append(lines, strings.Join(curLine, "\n"))
			curLine = curLine[:0]
			if len(word) > width {
				return s, fmt.Errorf("width (%d) is too small", width)
			}
			cursor = len(word)
			curLine = append(curLine, word)
		} else {
			cursor += len(word)
			curLine = append(curLine, word)
		}
	}
	if len(curLine) > 0 {
		lines = append(lines, strings.Join(curLine, "\n"))
	}
	return strings.Join(lines, "\n"), nil
}
