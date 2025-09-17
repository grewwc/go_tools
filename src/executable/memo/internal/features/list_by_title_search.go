package features

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func listByTitle(parser *terminalw.Parser) {
	title := parser.GetFlagValueDefault("title", "")
	if title == "" {
		title = parser.GetFlagValueDefault("c", "")
	}
	tags := []string{}
	records, _ := internal.ListRecords(internal.RecordLimit, internal.Reverse,
		internal.IncludeFinished, tags, parser.ContainsFlagStrict("and"), title, internal.Prefix)

	if parser.ContainsFlagStrict("count") {
		fmt.Printf("%d records found\n", len(records))
		return
	}
	if !parser.ContainsFlagStrict("out") && !internal.ToBinary {
		for _, record := range records {
			internal.PrintSeperator()
			p := regexp.MustCompile(`(?i)` + title)
			if !internal.Verbose {
				record.Title = "<hidden>"
			}
			internal.ColoringRecord(record, p)
			fmt.Println(record)
			fmt.Println(color.HiRedString(record.ID.String()))
		}
	} else if internal.ToBinary {
		panic("not supported")
	} else {
		var err error
		if (utilsw.IsExist(internal.TxtOutputName) && utilsw.PromptYesOrNo(fmt.Sprintf("%q already exists, do you want ot overwirte it? (y/n): ", internal.TxtOutputName))) ||
			!utilsw.IsExist(internal.TxtOutputName) {
			buf := bytes.NewBufferString("")
			for _, r := range records {
				buf.WriteString(fmt.Sprintf("%s %v %s\n", strings.Repeat("-", 10), r.Tags, strings.Repeat("-", 10)))
				buf.WriteString(r.Title)
				buf.WriteString("\n")
			}
			if err = os.WriteFile(internal.TxtOutputName, buf.Bytes(), 0666); err != nil {
				panic(err)
			}
		}
	}
}

func RegisterListByTitle(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return parser.ContainsFlagStrict("title") || parser.ContainsFlagStrict("c")
	}).Do(func() {
		listByTitle(parser)
	})
}
