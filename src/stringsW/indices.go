package stringsW

import (
	"strings"
)

func FindAll(str, substr string) []int {
	var result []int
	var index, newIndex int

	strLen, substrLen := len(str), len(substr)
	for index < strLen {
		newIndex = strings.Index(str[index:], substr)
		if newIndex == -1 {
			break
		}
		index += newIndex
		result = append(result, index)
		index += substrLen
	}
	return result
}

func StripPrefix(s, prefix string) string {
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return s
	}
	return s[idx+len(prefix):]
}

// SubStringQuiet
// beg include, end exclude
func SubStringQuiet(s string, beg, end int) string {
	if beg >= len(s) || end >= len(s) || beg >= end {
		return s
	}
	return s[beg:end]
}
