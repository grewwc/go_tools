package features

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
)

func RegisterCleanTag(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return p.ContainsFlagStrict("clean-tag")
	}).Do(func() {
		t := parser.GetFlagValueDefault("clean-tag", "")
		t = strings.ReplaceAll(t, ",", " ")
		tags := strw.SplitNoEmpty(t, " ")
		coloredTags := make([]string, len(tags))
		if len(tags) == 0 {
			fmt.Println("empty tags")
			return
		}
		for i := range tags {
			coloredTags[i] = color.HiRedString(tags[i])
		}
		fmt.Println("cleaning tags:", coloredTags)
		records, _ := internal.ListRecords(-1, false, true, tags, true, "", false)
		// fmt.Println("here", records)
		for _, record := range records {
			record.Delete()
		}
	})
}
