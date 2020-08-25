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
