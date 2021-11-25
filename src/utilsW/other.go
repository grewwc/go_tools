package utilsW

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func ToString(obj interface{}) string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	structName := fmt.Sprintf("%v {", t)

	buf := bytes.NewBufferString(structName)
	for i := 0; i < t.NumField(); i++ {
		if i > 0 {
			buf.WriteString(strings.Repeat(" ", len(structName)+2))
		} else {
			buf.WriteString("  ")
		}
		buf.WriteString(t.Field(i).Name)
		buf.WriteString(": ")
		var val string
		if v.Field(i).Type() == reflect.TypeOf(time.Time{}) {
			val = (v.Field(i).Interface().(time.Time)).Format("2006-01-02 15:04:05")
		} else {
			val = fmt.Sprintf("%v", v.Field(i))
		}
		buf.WriteString(val)
		if i+1 < t.NumField() {
			buf.WriteString("\n")
		}
	}
	buf.WriteString("  ")
	buf.WriteString("}")
	return buf.String()
}
