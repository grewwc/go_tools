package features

import (
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func insert(parser *terminalw.Parser) {
	internal.Insert(parser.CoExists("i", "e"), parser.GetFlagValueDefault("file", ""), "")
}

func RegisterInsert(parser *terminalw.Parser) {
	parser.On(func(r *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("i") || parser.CoExists("i", "e")
	}).Do(func() {
		insert(parser)
	})
}
