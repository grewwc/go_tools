package features

import (
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func delete(parser *terminalw.Parser) {
	arg := parser.GetFlagValueDefault("d", "")
	if !internal.DeleteRecordByTag(arg) {
		internal.DeleteRecord(arg, parser.ContainsFlagStrict("prev"))
	}
}

func RegisterDelete(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("d")
	}).Do(func() {
		delete(parser)
	})
}
