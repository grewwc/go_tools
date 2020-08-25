package test

import (
	"go_tools/src/terminalW"
	"testing"
)

func TestParser(t *testing.T) {
	res := terminalW.Parse("test.exe \"program dir\" -f file -a something night -v")
	t.Logf("positional : %v\n", res.Positional)

	t.Log(res.Optional)
}
