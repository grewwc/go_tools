package typew

import (
	"reflect"
	"unsafe"
)

// StrToBytes converts a string to []byte without copying.
func StrToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

// BytesToStr converts []byte to string without copying.
// This is safe because strings are immutable in Go.
func BytesToStr(b []byte) string {
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh := reflect.StringHeader{
		Data: sh.Data,
		Len:  sh.Len,
	}
	return *(*string)(unsafe.Pointer(&bh))
}
