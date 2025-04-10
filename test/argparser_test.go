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
	parser.ParseArgs("test.exe \"program dir\" -f file -a something night -v", "v")
	// test contains
	mustContain(t, parser, "v")
	mustContain(t, parser, "-v")
	mustContain(t, parser, "a")
	mustContain(t, parser, "f")

	// test positional args
	aim := cw.NewArrayList("program dir", "night")
	if !aim.Equals(parser.Positional) {
		t.Log(parser.Positional.ToStringSlice(), parser.Positional.Len())
		t.Log(aim.ToStringSlice(), aim.Size())
		t.Fail()
	}

	// test numbers

}
