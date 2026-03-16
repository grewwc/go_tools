package features

import (
	"fmt"

	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func pull(parser *terminalw.Parser) {
	fmt.Println("pulling...")
	host := parser.GetFlagValueDefault("host", internal.DefaultRemoteHost)
	if host == "" {
		panic("-pull requires --host <ip[:port]> or .configW:re.remote.host")
	}
	internal.SyncByHost(parser.GetFlagValueDefault("pull", ""), host, false, true)
}

func RegisterPull(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("pull")
	}).Do(func() {
		pull(parser)
	})
}
