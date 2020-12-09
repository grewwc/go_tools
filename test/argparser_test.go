package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
)

func mustContain(t *testing.T, parsedArgs *terminalW.ParsedResults, flag string) {
	if !parsedArgs.ContainsFlag(flag) {
		t.Fail()
	}
}
func TestParser(t *testing.T) {
	res := terminalW.ParseArgs("test.exe \"program dir\" -f file -a something night -v",
		"v")
	// test contains
	mustContain(t, res, "v")
	mustContain(t, res, "-v")
	mustContain(t, res, "a")
	mustContain(t, res, "f")

	// test positional args
	aim := containerW.NewSet()
	aim.AddAll("program dir", "night")
	if !aim.Equals(*res.Positional) {
		t.Log(res.Positional)
		t.Log(aim)
		t.Fail()
	}

	// test numbers

}
