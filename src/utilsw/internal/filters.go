package internal

import (
	"bytes"
	"strings"

	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/typesw"
)

type CommentsFilter struct {
}

func (f *CommentsFilter) Accept(b []byte) ([]byte, bool) {
	var buf bytes.Buffer
	var needHold bool

	for line := range strw.SplitByToken(bytes.NewReader(b), "\n", false) {
		needHold = false
		if (strings.Count(line, `"`)-strings.Count(line, `\"`))%2 == 1 {
			return b, true
		}
		parts := strw.SplitByStrKeepQuotes(line, "//", `"`, true)
		if len(parts) == 0 {
			continue
		}
		if len(parts) == 1 {
			trimed := bytes.TrimSpace(typesw.StrToBytes(parts[0]))
			if len(trimed) >= 2 && bytes.Equal(trimed[:2], []byte{'/', '/'}) {
				continue
			}
		}
		if len(parts) > 1 || !strings.Contains(parts[0], "//") {
			needHold = false
		} else {
			needHold = true
		}
		buf.WriteString(parts[0])
	}
	return buf.Bytes(), needHold
}

type SpecialCharsFilter struct {
}

func (f *SpecialCharsFilter) Accept(b []byte) ([]byte, bool) {
	var bb bytes.Buffer
	for _, r := range b {
		if r > 31 {
			bb.WriteByte(r)
		}
	}
	return bb.Bytes(), true
}
