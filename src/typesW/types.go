package typesW

import (
	"reflect"
	"strings"
)

type Comparable interface {
	Compare(any) int
}

type IntComparable int

func (i IntComparable) Compare(other any) int {
	return int(i - other.(IntComparable))
}

type CompareFunc[T any] func(a, b T) int

// CreateDefaultCmp is NOT Efficient.
func CreateDefaultCmp[T any]() CompareFunc[T] {
	var cmp CompareFunc[T]
	realType := reflect.TypeOf(*new(T))
	switch realType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cmp = func(a, b T) int {
			return int(reflect.ValueOf(a).Int()) - int(reflect.ValueOf(b).Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cmp = func(a, b T) int {
			return int(reflect.ValueOf(a).Uint()) - int(reflect.ValueOf(b).Uint())
		}
	case reflect.String:
		cmp = func(a, b T) int {
			return strings.Compare(reflect.ValueOf(a).String(), reflect.ValueOf(b).String())
		}
	}
	return cmp
}
