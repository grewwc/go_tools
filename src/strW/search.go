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
	matches := make([]int, 0, 2)
	next := kmpNext(pattern)
	j := 0
	// return []int{}
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
	n := len(pattern)
	next := make([]int, n) // 初始化 next 数组
	j := 0                 // j 指针指向当前最长前缀后缀匹配的位置

	for i := 1; i < n; i++ {
		// 当前字符不匹配时，回退到上一个匹配位置
		for j > 0 && pattern[i] != pattern[j] {
			j = next[j-1]
		}

		// 如果字符匹配，则继续扩展匹配长度
		if pattern[i] == pattern[j] {
			j++
		}

		// 更新 next 数组
		next[i] = j
	}
	return next
}
