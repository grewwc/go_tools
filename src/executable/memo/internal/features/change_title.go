package features

import (
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func changeTitle(parser *terminalw.Parser) {
	internal.ChangeTitle(parser.ContainsFlagStrict("file"),
		parser.CoExists("ct", "e"),
		parser.GetMultiFlagValDefault([]string{"ct", "cte", "ect"}, ""),
		parser.ContainsFlagStrict("prev"))
}

func RegisterChangeTitle(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("ct") || parser.CoExists("ct", "e")
	}).Do(func() {
		changeTitle(parser)
	})
}
