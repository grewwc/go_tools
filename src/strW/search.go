package strW

import (
	"strings"

	"github.com/grewwc/go_tools/src/typesW"
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
	matches := make([]int, 0)
	next := kmpNext(pattern)
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

func KmpSearchBytes(text, pattern []byte, n int) []int {
	if len(text) == 0 {
		return []int{}
	}
	if len(pattern) == 0 {
		return []int{0}
	}
	matches := make([]int, 0)
	next := kmpNext(typesW.BytesToStr(pattern))
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
	bStr := typesW.StrToBytes(str)
	bPattern := typesW.StrToBytes(pattern)
	indices := KmpSearchBytes(bStr, bPattern, 1)
	if len(indices) == 0 {
		return str
	}
	return typesW.BytesToStr(bStr[:indices[0]])
}

func TrimBefore(str, pattern string) string {
	bStr := typesW.StrToBytes(str)
	bPattern := typesW.StrToBytes(pattern)
	indices := KmpSearchBytes(bStr, bPattern, 1)
	if len(indices) == 0 {
		return str
	}
	return SubStringQuiet(str, GetLastItem[int](indices)+len(bPattern)+1, len(bStr))
}

func AllHasPrefix(str string, sub ...string) bool {
	for _, s := range sub {
		if !strings.HasPrefix(str, s) {
			return false
		}
	}
	return true
}

func kmpNext(pattern string) []int {
	next := make([]int, len(pattern))
	j := 0
	for i := 1; i < len(pattern); {
		if pattern[i] == pattern[j] {
			j++
			next[i] = j
			i++
		} else {
			if j == 0 {
				next[i] = 0
				i++
			} else {
				j = next[j-1]
			}
		}
	}
	return next
}
