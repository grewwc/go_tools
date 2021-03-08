package stringsW

import "strings"

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
	for idx, ch := range prefix {
		chStr := string(ch)
		if chStr != s[idx:idx+len(chStr)] {
			return s
		}
	}
	return s[len(prefix):]
}
