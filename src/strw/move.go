package strw

import (
	"strings"
)

func Move2EndAll(original, target string) string {
	n := strings.Count(original, target)
	if n <= 0 {
		return original
	}

	newString := strings.ReplaceAll(original, target, "")
	return newString + strings.Repeat(target, n)
}

func Reverse(str string) string {
	if len(str) == 0 {
		return str
	}
	builder := strings.Builder{}
	for i := len(str) - 1; i >= 0; i-- {
		builder.WriteByte(str[i])
	}
	return builder.String()
}
