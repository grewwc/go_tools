package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/utilsw"
)

func BenchmarkParseJson(b *testing.B) {
	fname := "bench_ordered_map.json"
	s := utilsw.ReadString(fname)
	for i := 0; i < b.N; i++ {
		utilsw.NewJsonFromString(s)
	}
}
