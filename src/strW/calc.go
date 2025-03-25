package strW

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var multiMinus = regexp.MustCompile(`(--)+`)
var positiveNumber = regexp.MustCompile(`\+(\d+)`)

func processInputStr(input string) string {
	input = multiMinus.ReplaceAllString(input, "+")
	input = positiveNumber.ReplaceAllString(input, "$1")
	return input
}

func plus(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	a, b = processInputStr(a), processInputStr(b)
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
	a, _ = removeLeadingZero(a)
	b, _ = removeLeadingZero(b)
	// a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
	// handle dot for float number
	n1, n2, numDot := countDotdigit(a, b, true)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	a += strings.Repeat("0", numDot-n1)
	b += strings.Repeat("0", numDot-n2)
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
	for idx+1 < len(res) && res[idx] == '0' {
		idx++
	}
	res = res[idx:]
	str := BytesToString(res)
	str = prependLeadingZero(str, numDot)
	str = removeSuffixZero(str)
	if isMinus {
		str = "-" + str
	}
	return str
}

func Plus(x ...string) string {
	if len(x) == 0 {
		return ""
	}
	res := x[0]
	for i := 1; i < len(x); i++ {
		res = plus(res, x[i])
	}
	return res
}

func Minus(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	a, b = processInputStr(a), processInputStr(b)
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
	if strings.HasPrefix(a, "-") && strings.HasPrefix(b, "-") {
		res := Minus(StripPrefix(a, "-"), StripPrefix(b, "-"))
		if len(res) > 0 && res[0] == '-' {
			return res[1:]
		}
		if res == "0" {
			return "0"
		}
		return "-" + res
	}

	// handle dot for float number
	n1, n2, numDot := countDotdigit(a, b, true)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	// a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
	isMinus := false
	a, b = StripPrefix(a, "-"), StripPrefix(b, "-")

	a, _ = removeLeadingZero(a)
	b, _ = removeLeadingZero(b)
	a += strings.Repeat("0", numDot-n1)
	b += strings.Repeat("0", numDot-n2)
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
	for idx+1 < len(res) && res[idx] == '0' {
		idx++
	}
	res = res[idx:]
	str := BytesToString(res)
	str = prependLeadingZero(str, numDot)
	str = removeSuffixZero(str)
	if isMinus {
		str = "-" + str
	}
	return str
}

func Neg(a string) string {
	if len(a) == 0 {
		return a
	}
	if a[0] == '-' {
		return SubStringQuiet(a, 1, len(a))
	}
	return fmt.Sprintf("-%s", a)
}

func mulInteger(s1, s2 string) string {
	var hold string = "0"
	m, n := len(s1), len(s2)
	res := make([]byte, m+n)
	for k := 0; k < m+n; k++ {
		for i := 0; i <= k; i++ {
			if i >= m {
				break
			}
			j := k - i
			if j >= n {
				continue
			}
			if j < 0 {
				break
			}
			tmp := int8(s1[m-1-i]-'0') * int8(s2[n-1-j]-'0')
			hold = Plus(hold, fmt.Sprintf("%d", tmp))
			// hold += int16(s1[m-1-i]-'0') * int16(s2[n-1-j]-'0')
		}
		// res[m+n-1-k] = byte(hold%10) + '0'
		res[m+n-1-k] = hold[len(hold)-1]
		// hold /= 10
		hold = hold[:len(hold)-1]
		if len(hold) == 0 {
			hold = "0"
		}
	}
	return BytesToString(res)
}

func mul(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	a, b = processInputStr(a), processInputStr(b)
	if len(a) == 0 || len(b) == 0 {
		return ""
	}
	isMinus := (a[0] == '-' && b[0] != '-') || (a[0] != '-' && b[0] == '-')
	a, b = StripPrefix(a, "-"), StripPrefix(b, "-")
	if len(a) > len(b) {
		a, b = b, a
	}
	if a == "0" || b == "0" {
		return "0"
	}

	// handle dot for float number
	_, _, numDot := countDotdigit(a, b, false)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	a, _ = removeLeadingZero(a)
	b, _ = removeLeadingZero(b)

	str := mulIntegerV2(a, b)
	str = prependLeadingZero(str, numDot)
	str = removeSuffixZero(str)
	str, _ = removeLeadingZero(str)
	if isMinus {
		str = "-" + str
	}

	return str
}

func order(s string) int {
	if len(s) == 0 {
		return 0
	}
	if s[0] == '-' {
		return len(s) - 1
	}
	return len(s)
}

func mulSingleDigit(a, b string) string {
	if a[0] == '-' && b[0] == '-' {
		res, _ := removeLeadingZero(mulInteger(a[1:], b[1:]))
		return res
	}
	if a[0] == '-' {
		res, _ := removeLeadingZero(mulInteger(a[1:], b))
		return "-" + res
	}
	if b[0] == '-' {
		res, _ := removeLeadingZero(mulInteger(a, b[1:]))
		return "-" + res
	}
	res, _ := removeLeadingZero(mulInteger(a, b))
	return res

}

func mulIntegerV2(x, y string) string {
	n := order(x)
	if n > order(y) {
		n = order(y)
	}
	if n == 0 || AnyEquals("0", x, y) {
		return "0"
	}
	if n == 1 {
		res := mulSingleDigit(x, y)
		// fmt.Println(x, y, res)
		return res
	}
	m := n / 2
	a := SubStringQuiet(x, 0, len(x)-m)
	if a == "" {
		a = "0"
	}
	a, _ = removeLeadingZero(a)
	b := SubStringQuiet(x, len(x)-m, len(x))
	if b == "" {
		b = x
	}
	b, _ = removeLeadingZero(b)
	c := SubStringQuiet(y, 0, len(y)-m)
	if c == "" {
		c = "0"
	}
	c, _ = removeLeadingZero(c)
	d := SubStringQuiet(y, len(y)-m, len(y))
	if d == "" {
		d = y
	}
	d, _ = removeLeadingZero(d)
	e := mulIntegerV2(a, c)
	f := mulIntegerV2(b, d)
	g := mulIntegerV2(Minus(a, b), Minus(c, d))
	tmp := Plus(e, f)
	tmp = Minus(tmp, g)
	m0 := strings.Repeat("0", m)
	return Plus(fmt.Sprintf("%s%s", e, strings.Repeat(m0, 2)), fmt.Sprintf("%s%s", tmp, m0), f)
}

func Mul(s ...string) string {
	if len(s) == 0 {
		return ""
	}
	res := s[0]
	for _, val := range s[1:] {
		res = mul(res, val)
	}
	return res
}

func Div(a, b string, numDigitToKeep int) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	a, b = processInputStr(a), processInputStr(b)
	if len(a) == 0 || len(b) == 0 {
		return ""
	}
	isMinus := (a[0] == '-' && b[0] != '-') || (a[0] != '-' && b[0] == '-')
	a, b = StripPrefix(a, "-"), StripPrefix(b, "-")
	d1, d2, d := countDotdigit(a, b, false)
	a, b = strings.ReplaceAll(a, ".", ""), strings.ReplaceAll(b, ".", "")
	// a, b = strings.TrimLeft(a, "0"), strings.TrimLeft(b, "0")
	a, _ = removeLeadingZero(a)
	b, _ = removeLeadingZero(b)
	if b == "0" {
		panic("b is 0")
	}
	if a == "0" {
		return "0"
	}
	// append '0' to a, b
	a += strings.Repeat("0", d-d1)
	b += strings.Repeat("0", d-d2)
	// <<< a, b are integers now
	decimalCount := 0
	res := ""

	aBak := a
	sumRemovedZero := 0
	var removedCount, prevRemoveCount int
	for decimalCount < numDigitToKeep+1 {
		divResult, addedZero, _ := Helper(&a, b)
		a, removedCount = removeLeadingZero(a)

		decimalCount += addedZero
		if addedZero > 1 {
			res += strings.Repeat("0", addedZero-1) + divResult
		} else if prevRemoveCount > 0 && addedZero > 0 {
			res += strings.Repeat("0", addedZero) + divResult
		} else {
			res += divResult
		}
		// fmt.Println("here", divResult, addedZero, a, removedCount, prevRemoveCount)
		sumRemovedZero += removedCount
		prevRemoveCount = removedCount
		if aBak == Mul(res, b) {
			break
		}
	}
	exp := exponent(aBak, b)
	// res = strings.TrimLeft(res, "0")
	res, _ = removeLeadingZero(res)
	decimalCount = len(res) - exp - 1

	// fmt.Println("decimalCount", decimalCount, len(res), res, exp)
	res = prependLeadingZero(res, decimalCount)
	if isMinus {
		res = "-" + res
	}
	res = round(res, numDigitToKeep)
	res = removeSuffixZero(res)

	return res
}

func Helper(a *string, b string) (string, int, bool) {
	if len((*a)) > len(b) {
		if (*a)[:len(b)] >= b {
			modified, divResult, clean := doDiv((*a)[:len(b)], b)
			*a = modified + (*a)[len(b):]
			return divResult, 0, clean
		} else {
			modified, divResult, clean := doDiv((*a)[:len(b)+1], b)
			*a = modified + (*a)[len(b)+1:]
			return divResult, 0, clean
		}
	} else if len((*a)) == len(b) {
		if (*a) >= b {
			modified, divResult, clean := doDiv((*a), b)
			*a = modified
			return divResult, 0, clean
		} else {
			(*a) += "0"
			modified, divResult, clean := doDiv((*a), b)
			*a = modified
			return divResult, 1, clean
		}
	} else {
		zeroCount := 0
		// prepend 0
		for len((*a)) < len(b) {
			*a += "0"
			zeroCount++
		}
		divResult, addedZero, clean := Helper(a, b)

		return divResult, addedZero + zeroCount, clean
	}
}

func exponent(a, b string) int {
	if len(a) == len(b) {
		if a >= b {
			return 0
		}
		return -1
	}
	cnt := len(a) - len(b)
	if cnt < 0 {
		a += strings.Repeat("0", -cnt)
		return cnt + exponent(a, b)
	}

	b += strings.Repeat("0", cnt)
	return cnt + exponent(a, b)
}

func doDiv(a, b string) (string, string, bool) {
	// assert len(a) == len(b) || len(a) == len(b) + 1
	if b == "1" {
		return "0", a, true
	}
	res := 0
	var curr = a
	for {
		val := Minus(curr, b)
		if val[0] != '-' || val == "0" {
			res++
			curr = val
		} else {
			break
		}
	}
	// fmt.Println("DoDiv", a, b, curr, res)
	return curr, strconv.Itoa(res), curr == "0"
}

func round(s string, digitToKeep int) string {
	idx := strings.LastIndexByte(s, '.')
	if idx < 0 {
		return s
	}
	if digitToKeep <= 0 {
		add := "0"
		if s[idx+1] >= '5' {
			add = "1"
		}
		return Plus(s[:idx], add)
	}
	numDigit := len(s) - idx - 1
	if numDigit <= digitToKeep {
		return s
	}
	val := s[idx+digitToKeep+1]
	if val < '5' {
		return s[:idx+digitToKeep+1]
	} else {
		// return Plus(s[:idx+digitToKeep+1], "0."+strings.Repeat("0", digitToKeep-1)+"1")
		return plus(s[:idx+digitToKeep+1], fmt.Sprintf("0.%s1", strings.Repeat("0", digitToKeep-1)))
	}
}
