package features

import (
	"fmt"

	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func push(parser *terminalw.Parser) {
	fmt.Println("pushing...")
	id := parser.GetFlagValueDefault("push", "")
	host := parser.GetFlagValueDefault("host", internal.DefaultRemoteHost)
	if host == "" {
		panic("-push requires --host <ip[:port]> or .configW:re.remote.host")
	}
	internal.SyncByIDToHost(id, host, true)
}

func RegisterPush(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("push") && parser.GetFlagValueDefault("push", "") != ""
	}).Do(func() {
		push(parser)
	})
}
