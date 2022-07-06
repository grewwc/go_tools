package utilsW

import (
	"time"
)

const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = DateFormat + "T" + TimeFormat + ".000Z"
)

func GetFirstDayOfThisWeek() time.Time {
	now := time.Now().Local()
	delta := int(now.Weekday()) - int(time.Monday)
	return now.AddDate(0, 0, -delta)
}
