package features

import (
	"fmt"

	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func push(parser *terminalw.Parser) {
	fmt.Println("pushing...")
	fmt.Println(parser.GetFlagValueDefault("push", ""))
	internal.SyncByID(parser.GetFlagValueDefault("push", ""), true, true)
}

func RegisterPush(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("push") && parser.GetFlagValueDefault("push", "") != ""
	}).Do(func() {
		push(parser)
	})
}
