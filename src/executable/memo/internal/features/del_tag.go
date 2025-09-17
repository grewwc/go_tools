package features

import (
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func delTag(parser *terminalw.Parser) {
	internal.AddTag(false, parser.GetFlagValueDefault("del-tag", ""), parser.ContainsFlagStrict("prev"))
}

func RegisterDelTag(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("del-tag")
	}).Do(func() {
		delTag(parser)
	})
}
