package stringsW

import (
	"fmt"
	"strings"
)

func Wrap(s string, width, indent int, delimiter string) string {
	words := SplitNoEmpty(s, " ")
	lines := make([]string, 0, 8)
	curLine := make([]string, 0, 8)
	cursor := 0

	for _, word := range words {
		word = strings.TrimSpace(word)
		if cursor+len(word) > width {
			// fmt.Println("what", len(strings.Join(curLine, dilimiter)), cursor, width)
			lines = append(lines, strings.TrimRight(strings.Join(curLine, ""), " "))
			curLine = curLine[:0]
			if len(word) > width {
				cursor = 0
				lines = append(lines, word)
				continue
			}
			cursor = len(word) + len(delimiter)
			curLine = []string{word + delimiter}
		} else {
			cursor += len(word) + len(delimiter)
			curLine = append(curLine, word+delimiter)
		}
	}
	if len(curLine) > 0 {
		lines = append(lines, strings.TrimRight(strings.Join(curLine, ""), " "))
	}
	res := ""
	for _, line := range lines {
		// fmt.Printf("|%s|\n", line)
		res += strings.Repeat(" ", indent) + line + "\n"
	}
	return strings.TrimRight(res, "\n")
}

func FormatInt64(val int64) string {
	suffix := ""
	var decimal int64
	var _1K int64 = 1 << 10
	_1M := _1K * _1K
	_1G := _1M * _1K
	_1T := _1G * _1K
	_1P := _1T * _1K

	if val < _1K {
		// suffix doesn't change
	} else if val < _1M {
		suffix = "K"
		decimal = val % _1K
		val /= _1K
	} else if val < _1G {
		suffix = "M"
		decimal = val % _1M
		val /= _1M
	} else if val < _1T {
		suffix = "G"
		decimal = val % _1G
		val /= _1G
	} else if val < _1P {
		suffix = "T"
		decimal = val % _1T
		val /= _1T
	} else {
		suffix = "P"
		decimal = val % _1P
		val /= _1P
	}
	if suffix != "" {
		return fmt.Sprintf("%v.%v%s", val, decimal, suffix)
	} else {
		return fmt.Sprintf("%v", val)
	}
}
