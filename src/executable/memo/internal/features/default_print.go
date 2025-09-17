package features

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
)

func printdefault(parser *terminalw.Parser) {
	records, _ := internal.ListRecords(internal.RecordLimit, false, false, []string{"todo", "urgent"}, false, "", true)
	for _, record := range records {
		internal.PrintSeperator()
		internal.ColoringRecord(record, nil)
		fmt.Println(record)
		fmt.Println(color.HiRedString(record.ID.String()))
	}
}
func RegisterDefault(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return p.Empty()
	}).Do(func() {
		printdefault(parser)
	})
}
