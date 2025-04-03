package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/algoW"
	"github.com/grewwc/go_tools/src/typesW"
)

func TestEditDistance(t *testing.T) {
	input := [][]string{
		{"abc", "abd"},
		{"hello", "hell"},
		{"cat", "cart"},
		{"kitten", "sitting"},
		{"abc", "def"},
		{"test", "tests"},
		{"example", "exmaple"},
		{"short", "shot"},
		{"apple", "appla"},
		{"intention", "execution"},
	}
	truth := []int{
		1, 1, 1, 3, 3, 1, 2, 1, 1, 5,
	}
	for i := 0; i < len(input); i++ {
		dist := algoW.EditDistance(typesW.StrToBytes(input[i][0]), typesW.StrToBytes(input[i][1]), nil)
		if dist != truth[i] {
			t.Errorf("Expect dist(%s, %s)=%d, but found: %d", input[i][0], input[i][1], truth[i], dist)
		}
	}
}
