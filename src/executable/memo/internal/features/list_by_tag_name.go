package features

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func printTitleWithColoredSeprator(title string) {
	colored := color.CyanString(strings.Repeat("~~~~~~~~~~~~", 10))
	title = strings.ReplaceAll(title, "<sep>", colored)
	fmt.Println(title)
}

func action(parser *terminalw.Parser) {
	tags := strw.SplitNoEmpty(strings.TrimSpace(parser.GetMultiFlagValDefault([]string{"t", "ta", "at"}, "")), " ")
	if parser.ContainsFlagStrict("pull") {
		internal.Remote.Set(true)
	}
	var records []*internal.Record
	// 如果是 id，特殊处理
	if internal.IsObjectID(parser.GetFlagValueDefault("t", "")) {
		id, err := primitive.ObjectIDFromHex(parser.GetFlagValueDefault("t", ""))
		if err != nil {
			panic(err)
		}
		r := &internal.Record{ID: id}
		r.LoadByID()
		if r.Invalid {
			return
		}
		records = []*internal.Record{r}
	} else {
		records, _ = internal.ListRecords(internal.RecordLimit, internal.Reverse, internal.IncludeFinished,
			tags, parser.ContainsFlagStrict("and"), "", parser.ContainsAnyFlagStrict("prefix", "pre", "a"))
	}
	if parser.ContainsFlagStrict("count") {
		fmt.Printf("%d records found\n", len(records))
		return
	}
	if !parser.ContainsAnyFlagStrict("pull", "push") {
		ignoreFields := []string{"AddDate", "ModifiedDate", "Invalid", "Title"}
		if internal.Verbose {
			ignoreFields = []string{}
		}
		// to stdout
		if !parser.ContainsFlagStrict("out") && !internal.ToBinary {
			if internal.OnlyTags {
				s := cw.NewTreeSet[string](nil)
				for _, r := range records {
					for _, t := range r.Tags {
						s.Add(t)
					}
				}
				for t := range s.Iterate() {
					fmt.Printf("%q  ", t)
				}
				fmt.Println()
			} else {
				for _, record := range records {
					internal.PrintSeperator()
					internal.ColoringRecord(record, nil)
					if !utilsw.IsText([]byte(record.Title)) {
						record.Title = color.HiYellowString("<binary>")
					}
					fmt.Println(utilsw.ToString(record, ignoreFields...))
					// print title
					printTitleWithColoredSeprator(record.Title)
					fmt.Println(color.HiRedString(record.ID.String()))
				}
			}
		} else if internal.ToBinary {
			for i := range records {
				content := records[i].Title
				idx := strings.IndexByte(content, '\n')
				filename := content[:idx]
				title := content[idx+1:]
				if !utilsw.IsExist(filename) ||
					(utilsw.IsExist(filename) && parser.ContainsFlagStrict("force")) ||
					(utilsw.IsExist(filename) && utilsw.PromptYesOrNo((fmt.Sprintf("%q already exists, do you want ot overwirte it? (y/n): ", filename)))) {
					if err := os.WriteFile(filename, []byte(title), 0666); err != nil {
						panic(err)
					}
				}
			}
		} else { // to file
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
	} else {
		wg := sync.WaitGroup{}
		wg.Add(len(records))
		for _, r := range records {
			go func(r *internal.Record) {
				fmt.Printf("begin to sync %s...\n", r.ID.Hex())
				internal.SyncByID(r.ID.Hex(), parser.ContainsFlagStrict("push"), true)
				fmt.Println("finished syncing")
				wg.Done()
			}(r)
		}
		utilsw.TimeoutWait(&wg, 30*time.Second)
	}
}

func RegisterListByTagName(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return (parser.ContainsFlagStrict("t") || parser.CoExists("t", "a")) && !internal.ListTagsAndOrderByTime
	}).Do(func() {
		action(parser)
	})
}
