package stringsW

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
	matches := KmpSearch(s, sub)
	return len(matches) > 0
}

func KmpSearch(text, pattern string) []int {
	matches := make([]int, 0)
	next := kmpPrefix(pattern)
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
		}
	}
	return matches
}

func CopySlice(original []string) []string {
	res := make([]string, 0, len(original))
	res = append(res, original...)
	return res
}

func EqualsAny(target string, choices ...string) bool {
	for _, choice := range choices {
		if choice == target {
			return true
		}
	}
	return false
}

func kmpPrefix(pattern string) []int {
	prefix := make([]int, len(pattern))
	j := 0
	for i := 1; i < len(pattern); {
		if pattern[i] == pattern[j] {
			j++
			prefix[i] = j
			i++
		} else {
			if j == 0 {
				prefix[i] = 0
				i++
			} else {
				j = prefix[j-1]
			}
		}
	}
	return prefix
}
