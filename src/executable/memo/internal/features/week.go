package features

import (
	"bytes"
	"fmt"
	"time"

	"github.com/grewwc/go_tools/src/executable/memo/internal"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func RegisterWeek(parser *terminalw.Parser) {
	positional := parser.Positional
	parser.On(func(p *terminalw.Parser) bool {
		return positional.Contains("week", nil)
	}).Do(func() {
		firstDay := utilsw.GetFirstDayOfThisWeek()
		now := time.Now()
		tag := firstDay.Format(fmt.Sprintf("%s.%s", "week", utilsw.DateFormat))
		rs, _ := internal.ListRecords(-1, true, true, []string{tag}, false, "", false)
		title := bytes.NewBufferString("")
		newWeekRecord := false
		if len(rs) > 1 {
			panic("too many week tags ")
		}
		if len(rs) == 0 {
			rs = []*internal.Record{internal.NewRecord("", tag)}
			newWeekRecord = true
		}
		for firstDay.Before(now) {
			dayTag := firstDay.Format(fmt.Sprintf("%s.%s", "log", utilsw.DateFormat))
			r, _ := internal.ListRecords(-1, true, true, []string{dayTag}, false, "", false)
			if len(r) > 1 {
				panic("log failed")
			}
			if len(r) == 1 {
				title.WriteString(fmt.Sprintf("-- %s --", firstDay.Format(utilsw.DateFormat)))
				title.WriteString("\n")
				title.WriteString(r[0].Title)
				title.WriteString("\n\n")
			}
			firstDay = firstDay.AddDate(0, 0, 1)
		}
		rs[0].Title = title.String()
		if newWeekRecord {
			rs[0].Save(true)
		} else {
			rs[0].Update(true)
		}
	})
}
