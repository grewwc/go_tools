package _utils_helpers

import (
	"fmt"
	"reflect"
	"strings"
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

func AddTopicToMethodName(topic, methodName string) string {
	return fmt.Sprintf("__%s__%s", topic, methodName)
}

func RemoveTopicFromMethodName(topic, methodName string) string {
	key := fmt.Sprintf("__%s__", topic)
	if !strings.HasPrefix(methodName, key) {
		return methodName
	}
	return strings.TrimPrefix(methodName, key)
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
