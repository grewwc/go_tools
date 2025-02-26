package _utils_helpers

import (
	"fmt"
	"reflect"
)

func GetMethods(obj interface{}) []*reflect.Method {
	t := reflect.TypeOf(obj)
	if t == nil {
		return nil
	}
	var result []*reflect.Method
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		if m.Type.NumIn() < 2 {
			continue
		}
		if m.Type.In(1).String() == "utilsW.Subscribe" {
			result = append(result, &m)
		}
	}
	return result
}

func MethodToString(m *reflect.Method) string {
	return fmt.Sprintf("%s_%s_%d", m.Type.String(), m.PkgPath, m.Index)
}

func MethodArrToString(methods []*reflect.Method) []string {
	result := make([]string, 0, len(methods))
	for _, m := range methods {
		result = append(result, MethodToString(m))
	}
	return result
}
