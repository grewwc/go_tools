package strW

import (
	"log"
	"strings"

	"github.com/grewwc/go_tools/src/algoW"
)

// countDotdigit count digit number of a, b, and max(a, b)
// e.g.: assert countDotdigit("0.3", "0.513", false) == 1,3,3
// e.g.: assert countDotdigit("0.3", "0.513", true) == 1,3,4
func countDotdigit(a, b string, add bool) (int, int, int) {
	aDot, bDot := strings.Count(a, "."), strings.Count(b, ".")
	if aDot > 1 || bDot > 1 {
		log.Fatalf("invalid number: %s, %s\n", a, b)
	}
	ai, bi := strings.IndexByte(a, '.'), strings.IndexByte(b, '.')
	if ai == -1 {
		ai = len(a) - 1
	}
	if bi == -1 {
		bi = len(b) - 1
	}
	ca := len(a) - ai - 1
	cb := len(b) - bi - 1
	// multiply
	if !add {
		return ca, cb, ca + cb
	}
	return ca, cb, algoW.Max(ca, cb)
}

func prependLeadingZero(str string, decimalCount int) string {
	if decimalCount > 0 {
		leading0 := false
		if len(str) <= decimalCount {
			str = strings.Repeat("0", decimalCount-len(str)) + str
			leading0 = true
		}
		str = str[:len(str)-decimalCount] + "." + str[len(str)-decimalCount:]
		if leading0 {
			str = "0" + str
		}
	}

	return str
}

func removeSuffixZero(str string) string {
	if strings.Count(str, ".") <= 0 {
		return str
	}
	// 删除最后的0
	str = strings.TrimRight(str, "0")
	if str[len(str)-1] == '.' {
		str = str[:len(str)-1]
	}
	return str
}

func removeLeadingZero(str string) (string, int) {
	if len(str) == 0 {
		return str, 0
	}
	idx := 0
	for idx+1 < len(str) && str[idx] == '0' {
		idx++
	}
	return str[idx:], idx
}
