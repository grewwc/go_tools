package features

import (
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func addTags(parser *terminalw.Parser) {
	internal.AddTag(true, parser.GetFlagValueDefault("add-tag", ""), parser.ContainsFlagStrict("prev"))
}

func RegisterAddTag(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("add-tag")
	}).Do(func() {
		addTags(parser)
	})
}
