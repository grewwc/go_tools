package strW

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
