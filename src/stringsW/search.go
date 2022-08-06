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
	if len(sub) == 0 {
		return len(s) != 0
	}
	prefix := makePrefix(s, sub)
	i, j := 0, 0
	for i < len(s) && j < len(sub) {
		if s[i] == sub[j] {
			i++
			j++
		} else {
			j = prefix[j] - 1
			if j == -1 {
				j = 0
				i++
			}
		}
	}
	return j == len(sub)
}

func CopySlice(original []string) []string {
	res := make([]string, 0, len(original))
	res = append(res, original...)
	return res
}

func makePrefix(s, sub string) []int {
	res := make([]int, len(sub))
	j, i := 0, 1
	for i < len(sub) {
		if sub[j] == sub[i] {
			res[i] = j + 1
			i++
			j++
		} else if j == 0 {
			res[i] = 0
			i++
		} else {
			j = res[j-1]
		}
	}
	return res
}
