package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/numW"
	"github.com/grewwc/go_tools/src/stringsW"
)

func TestFindAll(t *testing.T) {
	allString := "test.exe \"program dir\" -f file -a something night -v"
	substr := "something"
	result := stringsW.FindAll(allString, substr)
	t.Log(result)
}

func genRandomStrings(n int) string {
	allChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	indices := numW.RandInt(0, len(allChars), n)
	buf := make([]byte, len(indices))
	for i, idx := range indices {
		buf[i] = allChars[idx]
	}
	return stringsW.BytesToString(buf)
}

func TestContains(t *testing.T) {
	tests := []struct {
		s        string
		sub      string
		expected bool
	}{
		{"", "", false},
		{"", "a", false},
		{"a", "", true},
		{"abc", "b", true},
		{"abc", "d", false},
		{"abc", "abc", true},
		{"abc", "abcd", false},
		{"abcabc", "abc", true},
		{"abcabc", "bca", true},
		{"abcabc", "cab", true},
		{"abcabc", "dab", false},
		{"abcabc", "abcabc", true},
		{"abcabc", "abcabcabc", false},
		{"abcabc", "a", true},
		{"abcabc", "d", false},
	}

	for i := 0; i < 1000; i++ {
		target := genRandomStrings(100)
		prefix := genRandomStrings(1000)
		suffix := genRandomStrings(500)
		tests = append(tests, struct {
			s        string
			sub      string
			expected bool
		}{
			s:        prefix + target + suffix,
			sub:      target,
			expected: true,
		})
	}

	for _, test := range tests {
		result := stringsW.Contains(test.s, test.sub)
		if result != test.expected {
			t.Errorf("Contains(%q, %q) = %v, want %v", test.s, test.sub, result, test.expected)
		}
	}
}
