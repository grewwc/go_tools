package test

import (
	"go_tools/src/containerW"
	"log"
	"testing"
)

var s *containerW.Set
var inputs []string

func init() {
	s = containerW.NewSet()
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
