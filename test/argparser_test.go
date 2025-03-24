package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
)

func mustContain(t *testing.T, parsedArgs *terminalW.Parser, flag string) {
	if !parsedArgs.ContainsFlag(flag) {
		t.Fail()
	}
}
func TestParser(t *testing.T) {
	parser := terminalW.NewParser()
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
	aim := containerW.NewOrderedSet()
	aim.AddAll("program dir", "night")
	if !aim.Equals(*parser.Positional) {
		t.Log(parser.Positional)
		t.Log(aim)
		t.Fail()
	}

	// test numbers

}
