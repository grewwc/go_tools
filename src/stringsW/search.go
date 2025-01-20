package stringsW

import (
	"fmt"
	"math"
	"strings"
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
	matches := KmpSearch(s, sub)
	return len(matches) > 0
}

func KmpSearch(text, pattern string) []int {
	if len(text) == 0 {
		return []int{}
	}
	if len(pattern) == 0 {
		return []int{0}
	}
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

func Plus(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	if !strings.HasPrefix(a, "-") && strings.HasPrefix(b, "-") {
		return Minus(a, StripPrefix(b, "-"))
	}
	if strings.HasPrefix(a, "-") && !strings.HasPrefix(b, "-") {
		return Minus(b, StripPrefix(a, "-"))
	}
	isMinus := false
	// both have minus sign
	if strings.HasPrefix(a, "-") && strings.HasPrefix(b, "-") {
		a = StripPrefix(a, "-")
		b = StripPrefix(b, "-")
		isMinus = true
	}
	a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
	// handle dot for float number
	numDot := countDotdigit(a, b, true)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	if len(a) < len(b) {
		a, b = b, a
	}
	res := make([]byte, len(a)+1)
	carry := 0
	j := len(b) - 1
	i := len(a) - 1
	idx := len(res) - 1
	for idx >= 0 {
		valB := 0
		valA := 0
		if i >= 0 {
			valA = int(a[i] - '0')
			i--
		}
		if j >= 0 {
			valB = int(b[j] - '0')
			j--
		}
		val := valA + valB + carry
		if val >= 10 {
			carry = 1
			val -= 10
		} else {
			carry = 0
		}
		res[idx] = byte(val + '0')
		idx--
	}
	idx = 0
	for idx < len(res) && res[idx] == '0' {
		idx++
	}
	res = res[idx:]
	str := BytesToString(res)
	if isMinus {
		str = "-" + str
	}
	if numDot > 0 {
		leading0 := false
		str = strings.TrimRight(str, "0")
		if len(str) <= numDot {
			str = strings.Repeat("0", numDot-len(str)) + str
			leading0 = true
		}
		str = str[:len(str)-numDot] + "." + str[len(str)-numDot:]
		if leading0 {
			str = "0" + str
		}
	}
	return str
}

func Minus(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if len(a) == 0 {
		if strings.HasPrefix(b, "-") {
			return StripPrefix(b, "-")
		}
		return "-" + b
	}
	if len(b) == 0 {
		return a
	}
	if !strings.HasPrefix(a, "-") && strings.HasPrefix(b, "-") {
		return Plus(a, StripPrefix(b, "-"))
	}
	if strings.HasPrefix(a, "-") && !strings.HasPrefix(b, "-") {
		return "-" + Plus(StripPrefix(a, "-"), b)
	}
	// handle dot for float number
	numDot := countDotdigit(a, b, true)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	isMinus := false
	a, b = StripPrefix(a, "-"), StripPrefix(b, "-")
	if len(a) < len(b) || (len(a) == len(b) && a < b) {
		isMinus = true
		a, b = b, a
	}
	res := make([]byte, len(a))
	idx := len(res) - 1
	i := len(a) - 1
	j := len(b) - 1
	borrow := 0
	for idx >= 0 {
		valA := 0
		if i >= 0 {
			valA = int(a[i] - '0')
			i--
		}
		valB := 0
		if j >= 0 {
			valB = int(b[j] - '0')
			j--
		}
		val := valA - valB - borrow
		if val < 0 {
			val += 10
			borrow = 1
		} else {
			borrow = 0
		}
		res[idx] = byte(val + '0')
		idx--
	}
	idx = 0
	for idx < len(res) && res[idx] == '0' {
		idx++
	}
	res = res[idx:]
	str := BytesToString(res)
	if isMinus {
		str = "-" + str
	}
	if numDot > 0 {
		leading0 := false
		str = strings.TrimRight(str, "0")
		if len(str) <= numDot {
			str = strings.Repeat("0", numDot-len(str)) + str
			leading0 = true
		}
		str = str[:len(str)-numDot] + "." + str[len(str)-numDot:]
		if leading0 {
			str = "0" + str
		}
	}
	return str
}

func Mul(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if len(a) == 0 || len(b) == 0 {
		return ""
	}
	isMinus := (a[0] == '-' && b[0] != '-') || (a[0] != '-' && b[0] == '-')
	a, b = StripPrefix(a, "-"), StripPrefix(b, "-")
	if len(a) > len(b) {
		a, b = b, a
	}
	// handle dot for float number
	numDot := countDotdigit(a, b, false)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	carry := make([]int, len(a)+len(b)+1)
	res := make([]byte, len(a)+len(b)+1)

	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			idx := i + j + 2
			valA := int(a[i] - '0')
			valB := int(b[j] - '0')
			val := valA*valB + carry[idx] + int(res[idx])
			carry[idx] = 0
			if val >= 10 {
				carry[idx-1] += val / 10
				val %= 10
			}
			res[idx] = byte(val)
		}
	}
	for i := len(carry) - 1; i >= 0; i-- {
		res[i] += byte(carry[i])
		if res[i] >= 10 {
			carry[i-1] += int(res[i] / 10)
			res[i] %= 10
		}
	}
	idx := 0
	for idx < len(res) && res[idx] == 0 {
		idx++
	}
	for i := idx; i < len(res); i++ {
		res[i] += '0'
	}
	res = res[idx:]
	str := BytesToString(res)
	if isMinus {
		str = "-" + str
	}
	if numDot > 0 {
		leading0 := false
		str = strings.TrimRight(str, "0")
		if len(str) <= numDot {
			str = strings.Repeat("0", numDot-len(str)) + str
			leading0 = true
		}
		str = str[:len(str)-numDot] + "." + str[len(str)-numDot:]
		if leading0 {
			str = "0" + str
		}
	}
	return str
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

func countDotdigit(a, b string, add bool) int {
	aDot, bDot := strings.Count(a, "."), strings.Count(b, ".")
	if aDot > 1 || bDot > 1 {
		panic(fmt.Sprintf("invalid number: %s, %s\n", a, b))
	}
	ai, bi := strings.IndexByte(a, '.'), strings.IndexByte(b, '.')
	if ai == -1 {
		ai = len(a) - 1
	}
	if bi == -1 {
		bi = len(b) - 1
	}
	// multiply
	if !add {
		return (len(a) - ai - 1) + (len(b) - bi - 1)
	}
	return int(math.Max(float64(len(a)-ai-1), float64(len(b)-bi-1)))
}
