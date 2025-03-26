package _utils_helpers

import (
	"bytes"
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/grewwc/go_tools/src/typesW"
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

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
func AddTopicToMethodName(topic, methodName string) string {
	return fmt.Sprintf("__%s__%s", topic, methodName)
}

func RemoveTopicFromMethodName(topic, methodName string) string {
	key := fmt.Sprintf("__%s__", topic)
	if !strings.HasPrefix(methodName, key) {
		return methodName
	}
	b := typesW.StringToBytes(methodName)
	return typesW.BytesToString(bytes.TrimPrefix(b, typesW.StringToBytes(key)))
}

func InterfaceToValue(args ...interface{}) []reflect.Value {
	result := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		result = append(result, reflect.ValueOf(arg))
	}
	return result
}

func MethodToString(topic string, m *reflect.Method) string {
	methodName := fmt.Sprintf("%s_%s_%d", m.Type.String(), m.PkgPath, m.Index)
	return AddTopicToMethodName(topic, methodName)
}

func MethodArrToString(topic string, methods []*reflect.Method) []string {
	result := make([]string, 0, len(methods))
	for _, m := range methods {
		result = append(result, MethodToString(topic, m))
	}
	return result
}
