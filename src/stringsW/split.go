package stringsW

import (
	"bytes"
	"io"
	"log"
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

func SplitByToken(reader io.Reader, token string, keepToken bool) <-chan string {
	if len(token) == 0 {
		panic("token should not be empty")
	}
	ch := make(chan string)
	tokenBytes := StringToBytes(token)
	go func() {
		defer close(ch)
		if reader == nil {
			return
		}
		var buf bytes.Buffer
		b := make([]byte, 1)
		curr := make([]byte, 0, len(tokenBytes))
		for {
			n, err := reader.Read(b)
			if n <= 0 {
				break
			}
			if err != nil && err != io.EOF {
				log.Fatal(err)
				break
			}
			end := (err == io.EOF)
			if len(curr) < len(tokenBytes) {
				curr = append(curr, b[0])
			} else {
				curr = curr[1:]
				curr = append(curr, b[0])
			}
			if _, err := buf.Write(b); err != nil {
				panic(err)
			}
			if bytes.Equal(curr, tokenBytes) {
				buf.Truncate(buf.Len() - len(tokenBytes))
				if buf.String() != "" {
					ch <- buf.String()
					buf.Reset()
				}
				if keepToken {
					if _, err := buf.Write(curr); err != nil {
						panic(err)
					}
				}
				curr = make([]byte, 0, len(tokenBytes))
			}
			if end {
				break
			}
		}

		if buf.Len() > 0 {
			ch <- buf.String()
		}
	}()
	return ch
}
