package test

import (
	"log"
	"testing"

	"github.com/grewwc/go_tools/src/cw"
)

var s *cw.Set
var inputs []string

func init() {
	s = cw.NewSet()
	inputs = []string{
		"good",
		"something",
		"what",
		"",
		"\n",
	}
}

func TestSetAdd(t *testing.T) {

	for _, input := range inputs {
		s.Add(input)
	}

	if s.Size() != len(inputs) {
		log.Fatal("size is wrong", s.Size(), len(inputs))
	}

	for _, input := range inputs {
		if !s.Contains(input) {
			t.Fatal("not exist", input)
		}
	}
}

func TestSetDelete(t *testing.T) {
	for _, input := range inputs {
		s.Delete(input)
		t.Log("deleting: ", input)
	}

	if !s.Empty() {
		t.Fatal("delete wrong")
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s.Add(12)
	}
}

func BenchmarkContains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s.Contains("123")
	}
}
