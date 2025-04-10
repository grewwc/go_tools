package strw

import (
	"strings"

	"github.com/grewwc/go_tools/src/algow"
)

func FindAll(str, substr string) []int {
	return KmpSearch(str, substr, -1)
}

// StripPrefix: strip prefix
//
// Deprecated: use strings.TrimPrefix instead
func StripPrefix(s, prefix string) string {
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return s
	}
	return s[idx+len(prefix):]
}

// StripSuffix strip suffix
//
// Deprecated: use strings.TrimSuffix instead
func StripSuffix(s, suffix string) string {
	idx := strings.LastIndex(s, suffix)
	if idx < 0 {
		return s
	}
	return s[:idx]
}

// SubStringQuiet
// beg include, end exclude
func SubStringQuiet(s string, beg, end int) string {
	if beg >= len(s) || beg >= end {
		return ""
	}
	beg = algow.Max(beg, 0)
	end = algow.Min(end, len(s))
	return s[beg:end]
}
