package features

import (
	"strconv"
	"time"

	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func RegisterLog(parser *terminalw.Parser) {
	positional := parser.Positional
	parser.On(func(p *terminalw.Parser) bool {
		return positional.Contains("log", nil)
	}).Do(func() {
		positional.Delete("log", nil)
		nextDay := 0
		var err error
		if positional.Len() == 1 {
			if nextDay, err = strconv.Atoi(positional.ToStringSlice()[0]); err != nil {
				nextDay = 0
			}
		}

		tag := time.Now().Add(time.Duration(nextDay * int(time.Hour) * 24)).Format("log.2006-01-02")
		rs, _ := internal.ListRecords(-1, true, true, []string{tag}, false, "", false)
		if len(rs) > 1 {
			panic("log failed: ")
		}
		if len(rs) == 0 {
			internal.Insert(true, "", tag)
		} else {
			parser.Optional.Put("-u", rs[0].ID.Hex())
			internal.Update(parser, false, true, false)
		}
	})
}
