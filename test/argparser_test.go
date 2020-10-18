package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/terminalW"
)

func TestParser(t *testing.T) {
	res := terminalW.ParseArgsCmd("test.exe \"program dir\" -f file -a something night -v")
	t.Logf("positional : %v\n", res.Positional)

	t.Log(res.Optional)
}
