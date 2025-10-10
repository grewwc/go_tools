package strw

import (
	"strings"

	"github.com/grewwc/go_tools/src/typesw"
)

// if target is in slice, return true
// else return false
func SliceContains(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

// Contains test if "sub" is a substring of s
func Contains(s, sub string) bool {
	if len(s) == 0 {
		return false
	}
	if len(sub) == 0 {
		return true
	}
	matches := KmpSearch(s, sub, 1)
	return len(matches) > 0
}

func KmpSearch(text, pattern string, n int) []int {
	if len(text) == 0 {
		return []int{}
	}
	if len(pattern) == 0 {
		return []int{0}
	}
	matches := make([]int, 0, 2)
	next := kmpNext(pattern)
	i := 0
	j := 0
	for i < len(text) {
		if text[i] == pattern[j] {
			i++
			j++
		} else if j == 0 {
			i++
		} else {
			j = next[j-1]
		}
		if j == len(pattern) {
			matches = append(matches, i-j)
			j = next[j-1]
			if n > 0 && len(matches) >= n {
				break
			}
		}
	}
	return matches
}

func KmpSearchBytes(text, pattern []byte, n int) []int {
	if len(text) == 0 {
		return []int{}
	}
	if len(pattern) == 0 {
		return []int{0}
	}
	matches := make([]int, 0)
	next := kmpNext(typesw.BytesToStr(pattern))
	j := 0
	for i := 0; i < len(text); {
		if text[i] == pattern[j] {
			i++
			j++
		} else {
			if j == 0 {
				i++
			} else {
				j = next[j-1]
			}
		}
		if j == len(pattern) {
			matches = append(matches, i-j)
			j = next[j-1]
			if n > 0 && len(matches) >= n {
				break
			}
		}
	}
	return matches
}

func CopySlice(original []string) []string {
	res := make([]string, 0, len(original))
	res = append(res, original...)
	return res
}

func AnyEquals(target string, choices ...string) bool {
	for _, choice := range choices {
		if choice == target {
			return true
		}
	}
	return false
}

func AnyContains(str string, searchStrings ...string) bool {
	for _, searchStr := range searchStrings {
		if Contains(str, searchStr) {
			return true
		}
	}
	return false
}

func AllContains(str string, searchStrings ...string) bool {
	for _, searchStr := range searchStrings {
		if !Contains(str, searchStr) {
			return false
		}
	}
	return true
}

func AnyHasPrefix(str string, sub ...string) bool {
	for _, s := range sub {
		if strings.HasPrefix(str, s) {
			return true
		}
	}
	return false
}

func TrimAfter(str, pattern string) string {
	bStr := typesw.StrToBytes(str)
	bPattern := typesw.StrToBytes(pattern)
	indices := KmpSearchBytes(bStr, bPattern, 1)
	if len(indices) == 0 {
		return str
	}
	return typesw.BytesToStr(bStr[:indices[0]])
}

func TrimBefore(str, pattern string) string {
	bStr := typesw.StrToBytes(str)
	bPattern := typesw.StrToBytes(pattern)
	indices := KmpSearchBytes(bStr, bPattern, 1)
	if len(indices) == 0 {
		return str
	}
	return SubStringQuiet(str, GetLastItem[int](indices)+len(bPattern)+1, len(bStr))
}

func kmpNext(pattern string) []int {
	next := make([]int, len(pattern))
	prefixLen := 0
	i := 1
	for i < len(pattern) {
		if pattern[i] == pattern[prefixLen] {
			prefixLen++
			next[i] = prefixLen
			i++
		} else if prefixLen == 0 {
			next[i] = 0
			i++
		} else {
			prefixLen = next[prefixLen-1]
		}
	}
	return next
}
