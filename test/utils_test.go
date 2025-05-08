package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/utilsw"
)

func TestGetCurrentFileName(t *testing.T) {
	expectedCurFilename := "utils_test.go"
	real := utilsw.GetCurrentFileName()
	if expectedCurFilename != real {
		t.Errorf("%q != %q\n", expectedCurFilename, real)
	}
}

func TestTrimFileExt(t *testing.T) {
	filename := "utils.test.go"
	expected := "utils.test"
	real := utilsw.TrimFileExt(filename)
	if expected != real {
		t.Errorf("%q != %q\n", expected, real)
	}
}
