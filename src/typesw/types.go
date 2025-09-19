package typesw

import (
	"hash/fnv"
	"reflect"
	"strings"
)

type CompareFunc[T any] func(a, b T) int
type HashFunc[T any] func(obj T) int

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
	case reflect.Float32, reflect.Float64:
		cmp = func(a, b T) int {
			diff := reflect.ValueOf(a).Float() - reflect.ValueOf(b).Float()
			if diff < 0 {
				return -1
			}
			if diff > 0 {
				return 1
			}
			return 0
		}
	case reflect.String:
		cmp = func(a, b T) int {
			return strings.Compare(reflect.ValueOf(a).String(), reflect.ValueOf(b).String())
		}
	case reflect.Pointer:
		cmp = func(a, b T) int {
			return int(reflect.ValueOf(a).Pointer()) - int(reflect.ValueOf(b).Pointer())
		}
	case reflect.Func:
		cmp = func(a, b T) int {
			return int(reflect.ValueOf(a).Pointer()) - int(reflect.ValueOf(b).Pointer())
		}
	}
	return cmp
}

func CreateDefaultHash[T any]() HashFunc[T] {
	var cmp HashFunc[T]
	val := *new(T)
	realType := reflect.TypeOf(val)
	switch realType.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		cmp = func(a T) int {
			return int(reflect.ValueOf(a).Int())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		cmp = func(a T) int {
			return int(reflect.ValueOf(a).Uint())
		}
	case reflect.String:
		cmp = func(a T) int {
			h := fnv.New32a()
			h.Write(StrToBytes(reflect.ValueOf(a).String()))
			return int(h.Sum32())
		}
	case reflect.Pointer:
		cmp = func(a T) int {
			return int(reflect.ValueOf(a).Pointer())
		}
	case reflect.Func:
		cmp = func(a T) int {
			return int(reflect.ValueOf(a).Pointer())
		}
	}
	return cmp
}
