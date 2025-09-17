package features

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func update(parser *terminalw.Parser) {
	positional := parser.Positional
	if parser.ContainsFlagStrict("u") || positional.Contains("u", nil) {
		positional.Delete("u", nil)
		var id string
		tags := positional.ToStringSlice()
		isObjectID := false
		if positional.Len() > 0 {
			isObjectID = internal.IsObjectID(tags[0])
		}
		// tags 里面可能是 objectid
		if len(tags) == 1 && isObjectID {
			id = tags[0]
			goto tagIsId
		}

		if len(tags) > 0 {
			if r, _ := internal.ListRecords(-1, true, false, tags, false, "", internal.Prefix); len(r) < 1 {
				fmt.Println(color.YellowString("no records associated with the tags (%v: prefix: %v) found", tags, internal.Prefix))
				return
			}
		}
		id = internal.ReadInfo(false)

	tagIsId:
		parser.Optional.Put("-u", id)
		if id != "" {
			parser.Optional.Put("-e", "")
		}
		internal.Update(parser, parser.ContainsFlagStrict("file"), parser.ContainsFlagStrict("e"), id == "")
		return
	}
}

func RegisterUpdate(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("u") || p.Positional.Contains("u", nil)
	}).Do(func() {
		update(parser)
	})
}
