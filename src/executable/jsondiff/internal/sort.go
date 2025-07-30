package internal

import "encoding/json"

func compareString(s1, s2 string) int {
	if s1 < s2 {
		return -1
	}
	if s1 == s2 {
		return 0
	}
	return 1
}

func convertToString(a any) string {
	var s string
	var n json.Number
	var ok bool
	n, ok = a.(json.Number)
	if !ok {
		s, ok = a.(string)
		if !ok {
			return ""
		} else {
			return s
		}
	}
	return string(n)
}

func SortJson(a, b any) int {
	return compareString(convertToString(a), convertToString(b))
}
