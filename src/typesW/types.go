package typesW

import (
	"reflect"
	"strings"
)

type Comparable interface {
	Compare(interface{}) int
}

type IntComparable int

func (i IntComparable) Compare(other interface{}) int {
	return int(i - other.(IntComparable))
}

type CompareFunc = func(a, b interface{}) int

func CreateDefaultCmp[T any]() CompareFunc {
	var cmp CompareFunc
	realType := reflect.TypeOf(*new(T))
	switch realType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cmp = func(a, b interface{}) int {
			return a.(int) - b.(int)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cmp = func(a, b interface{}) int {
			return int(a.(uint) - b.(uint))
		}
	case reflect.String:
		cmp = func(a, b interface{}) int {
			return strings.Compare(a.(string), b.(string))
		}
	}
	return cmp
}
