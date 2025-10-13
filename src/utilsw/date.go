package utilsw

import (
	"fmt"
	"time"
)

const (
	DateFormat         = "2006-01-02"
	TimeFormat         = "15:04:05"
	DateTimeFormatFull = DateFormat + "T" + TimeFormat + ".000Z"
	DateTimeFormat     = DateFormat + " " + TimeFormat
)

func GetFirstDayOfThisWeek() time.Time {
	now := time.Now().Local()
	delta := int(now.Weekday()) - int(time.Monday)
	if delta < 0 {
		delta += 7
	}

	return now.AddDate(0, 0, -delta)
}

func ToUnix(timeVal string) int {
	t, err := time.Parse(fmt.Sprintf("%s %s", DateFormat, TimeFormat), timeVal)
	if err != nil {
		t, err = time.Parse(DateFormat, timeVal)
		if err != nil {
			panic(err)
		}
	}
	return int(t.Unix())
}
