package strW

import "unicode"

func IsBlank(str string) bool {
	if str == "" {
		return true
	}
	for _, ch := range str {
		if !unicode.IsSpace(ch) {
			return false
		}
	}
	return true
}

func IsAllBlank(strs ...string) bool {
	if len(strs) == 0 {
		return false
	}
	for _, str := range strs {
		if !IsBlank(str) {
			return false
		}
	}
	return true
}

func IsAnyBlank(strs ...string) bool {
	for _, str := range strs {
		if IsBlank(str) {
			return true
		}
	}
	return false
}
