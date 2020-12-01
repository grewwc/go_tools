package stringsW

import (
	"fmt"
	"strings"
)

func Wrap(s string, width, indent int, dilimiter string) (string, error) {
	words := SplitNoEmpty(s, " ")
	lines := make([]string, 0, 8)
	curLine := make([]string, 0, 8)
	cursor := 0

	for _, word := range words {
		word = strings.TrimSpace(word)
		if cursor+len(word) > width {
			lines = append(lines, strings.Join(curLine, dilimiter))
			curLine = curLine[:0]
			if len(word) > width {
				return s, fmt.Errorf("width (%d) is too small", width)
			}
			cursor = len(word) + indent
			curLine = append(curLine, word+dilimiter)
		} else {
			cursor += len(word) + len(dilimiter)
			curLine = append(curLine, word+dilimiter)
		}
	}
	if len(curLine) > 0 {
		lines = append(lines, strings.Join(curLine, dilimiter))
	}
	res := ""
	for _, line := range lines {
		// fmt.Printf("|%s|\n", line)
		res += strings.Repeat(" ", indent) + line + "\n"
	}
	return res, nil
}
