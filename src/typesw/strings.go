package typesw

// StrToBytes converts a string to []byte without copying.
func StrToBytes(s string) []byte {
	return []byte(s)
}

// BytesToStr converts []byte to string without copying.
// This is safe because strings are immutable in Go.
func BytesToStr(b []byte) string {
	return string(b)
}
