package features

import (
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

var finish = internal.Finish

func RegisterNf(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return p.ContainsFlagStrict("nf")
	}).Do(func() {
		tags := []string{parser.GetFlagValueDefault("nf", "")}
		tags = append(tags, parser.GetPositionalArgs(true)...)
		for _, tag := range tags {
			internal.Toggle(false, internal.GetObjectIdByTags([]string{tag}, true), finish, parser.ContainsFlagStrict("prev"))
		}
	})
}

func RegisterF(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return p.ContainsFlagStrict("f")
	}).Do(func() {
		tags := []string{parser.GetFlagValueDefault("f", "")}
		tags = append(tags, parser.GetPositionalArgs(true)...)

		if internal.Prefix {
			for _, tag := range tags {
				internal.FinishRecordsByTags([]string{tag})
			}
			return
		}
		for _, tag := range tags {
			internal.Toggle(true, internal.GetObjectIdByTags([]string{tag}, false), finish, parser.ContainsFlagStrict("prev"))
		}

	})
}
