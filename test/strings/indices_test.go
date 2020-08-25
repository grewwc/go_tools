package strings

import (
	"github.com/grewwc/go_tools/src/stringsW"
	"testing"
)

func TestFindAll(t *testing.T) {
	allString := "test.exe \"program dir\" -f file -a something night -v"
	substr := "something"
	result := stringsW.FindAll(allString, substr)
	t.Log(result)
}
