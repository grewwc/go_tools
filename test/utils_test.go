package test

import (
	"testing"

	"github.com/grewwc/go_tools/src/utilsW"
)

func TestGetCurrentFileName(t *testing.T) {
	expectedCurFilename := "utils_test.go"
	real := utilsW.GetCurrentFileName()
	if expectedCurFilename != real {
		t.Errorf("%q != %q\n", expectedCurFilename, real)
	}
}

func TestTrimFileExt(t *testing.T) {
	filename := "utils.test.go"
	expected := "utils.test"
	real := utilsW.TrimFileExt(filename)
	if expected != real {
		t.Errorf("%q != %q\n", expected, real)
	}
}
