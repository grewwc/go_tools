package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/terminalw"
)

func mustContain(t *testing.T, parsedArgs *terminalw.Parser, flag string) {
	if !parsedArgs.ContainsFlag(flag) {
		t.Fail()
	}
}
func TestParser(t *testing.T) {
	parser := terminalw.NewParser()
	parser.Bool("v", false, "")
	parser.String("f", "", "")
	parser.Bool("a", false, "")
	parser.ParseArgs(`-f file -a something night -v`)
	// test contains
	mustContain(t, parser, "v")
	mustContain(t, parser, "-v")
	mustContain(t, parser, "a")
	mustContain(t, parser, "f")

	// test positional args
	aim := cw.NewLinkedList(`something`, "night")
	if !aim.Equals(parser.Positional, nil) {
		t.Log("parser.Positional:\n", parser.Positional.ToStringSlice(), parser.Positional.Len())
		t.Log("expected:\n", aim.ToStringSlice(), aim.Len())
		t.Fail()
	}

	// test numbers

}
