package utilsW

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"time"
	"unsafe"

	"github.com/grewwc/go_tools/src/containerW"
)

func toString(numTab int, obj interface{}, ignoresFieldName ...string) string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", obj)
	}
	copyV := reflect.New(v.Type()).Elem()
	copyV.Set(v)
	structName := fmt.Sprintf("%v {", t)
	s := containerW.NewSet()
	for _, ignore := range ignoresFieldName {
		s.Add(ignore)
	}
	first := true
	buf := bytes.NewBufferString(structName)
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if s.Contains(fieldName) {
			continue
		}
		if !first {
			buf.WriteString(strings.Repeat(" ", len(structName)+1+numTab))
		} else {
			first = false
			buf.WriteString(" ")
		}
		buf.WriteString(fieldName)
		buf.WriteString(": ")
		var val string
		fieldVal := copyV.Field(i)
		fieldVal = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
			val = (fieldVal.Interface().(time.Time)).Format("2006-01-02/15:04:05")
		} else {
			// val = fmt.Sprintf("%v", v.Field(i))
			val = toString(len(structName)+len(fieldName)+3, fieldVal.Interface())
		}
		buf.WriteString(val)
		buf.WriteString("\n")
	}
	buf.WriteString(strings.Repeat(" ", numTab))
	buf.WriteString("}")
	return buf.String()
}

func ToString(obj interface{}, ignoresFieldName ...string) string {
	return toString(0, obj, ignoresFieldName...)
}
