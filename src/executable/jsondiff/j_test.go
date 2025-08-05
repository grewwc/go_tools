package main

import (
	"testing"

	"github.com/grewwc/go_tools/src/utilsw"
)

func BenchmarkJsonDiff(b *testing.B) {
	mt = true
	j, _ := utilsw.NewJsonFromFile("/Users/wwc129/self-dev/main-test/1.json")
	for i := 0; i < b.N; i++ {
		compareJson("", j, j)
	}
}
