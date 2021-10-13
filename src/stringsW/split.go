package stringsW

import (
	"bytes"
	"strings"
)

// SplitNoEmpty remove empty strings
func SplitNoEmpty(str, sep string) []string {
	var res []string
	if str == "" {
		return res
	}
	for _, s := range strings.Split(str, sep) {
		if s == "" {
			continue
		}
		res = append(res, s)
	}
	return res
}

// SplitNoEmptyKeepQuote keep content in quote intact
func SplitNoEmptyKeepQuote(str string, sep rune) []string {
	inQuote := false
	var res []string
	var word bytes.Buffer
	if str == "" {
		return res
	}

	for _, s := range str {
		if s == '"' {
			inQuote = !inQuote
		} else if s != sep || inQuote {
			word.WriteRune(s)
		} else if word.Len() != 0 {
			res = append(res, word.String())
			word.Reset()
		}
	}
	if word.Len() != 0 {
		res = append(res, word.String())
	}
	return res
}

func ReplaceAllInQuoteUnchange(s string, old, new rune) string {
	inQuote := false
	var res bytes.Buffer
	for _, ch := range s {
		if ch == '"' {
			inQuote = !inQuote
			res.WriteRune(ch)
			continue
		}
		if ch == old {
			if inQuote {
				res.WriteRune(old)
			} else {
				res.WriteRune(new)
			}
		} else {
			res.WriteRune(ch)
		}
	}
	return res.String()
}

func ReplaceAllOutQuoteUnchange(s string, old, new rune) string {
	inQuote := false
	var res bytes.Buffer
	for _, ch := range s {
		if ch == '"' {
			inQuote = !inQuote
			res.WriteRune(ch)
			continue
		}
		if ch == old {
			if inQuote {
				res.WriteRune(new)
			} else {
				res.WriteRune(old)
			}
		} else {
			res.WriteRune(ch)
		}
	}
	return res.String()
}

func GetLastItem(slice []string) string {
	if len(slice) < 1 {
		return ""
	}
	return slice[len(slice)-1]
}
