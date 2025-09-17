package features

import (
	"fmt"
	"strings"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func RegisterMove(parser *terminalw.Parser) {
	positional := parser.Positional
	parser.On(func(p *terminalw.Parser) bool {
		return positional.Contains("move", nil)
	}).Do(func() {
		s := positional.ToStringSlice()
		if len(s) != 3 {
			fmt.Println(">> re move absFileName type")
			return
		}
		type_, filename := s[2], s[1]
		logMsg := internal.LogMoveImages(type_, strings.ReplaceAll(filename, "\\\\", "\\"))
		tag := "move_" + type_
		rs, _ := internal.ListRecords(-1, true, true, []string{tag}, false, "", false)
		if len(rs) == 0 {
			internal.NewRecord(logMsg, tag).Save(false)
		} else {
			s := cw.NewOrderedSet()
			for _, title := range strings.Split(rs[0].Title, "\n") {
				s.Add(title)
			}
			for _, title := range strings.Split(logMsg, "\n") {
				s.Add(title)
			}
			rs[0].Title = strings.Join(s.ToStringSlice(), "\n")
			rs[0].Update(false)
		}
	})
}
