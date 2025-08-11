package strw

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strings"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/typesw"
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

// SplitNoEmptyPreserveQuote keep content in quote intact
func SplitNoEmptyPreserveQuote(str string, sep rune, symbols string, keepSymbol bool) []string {
	inQuote := false
	var res []string
	var word strings.Builder
	if str == "" {
		return res
	}
	s := cw.NewSetT[rune]()
	for _, symbol := range symbols {
		s.Add(symbol)
	}
	for _, ch := range str {
		if s.Contains(ch) {
			inQuote = !inQuote
			if keepSymbol {
				word.WriteRune(ch)
			}
		} else if ch != sep || inQuote {
			word.WriteRune(ch)
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

// SplitByStrKeepQuote keep content in quote intact
func SplitByStrKeepQuote(str string, sep string) []string {
	var res []string
	sepBytes := typesw.StrToBytes(sep)
	// var buf bytes.Buffer
	inquote := false
	var buf bytes.Buffer
	// var curr bytes.Buffer
	var prev byte
	for i, r := range str {
		if i > 0 {
			prev = str[i-1]
		}
		if r == '"' && prev != '\\' {
			inquote = !inquote
			buf.WriteRune(r)
		} else {
			if buf.Len() > len(sep) && bytes.Equal(buf.Bytes()[buf.Len()-len(sep):], sepBytes) {
				if inquote {
					buf.WriteRune(r)
				} else {
					content := buf.String()[:buf.Len()-len(sep)]
					if content != "" {
						res = append(res, content)
					}
					buf.Reset()
				}
			} else {
				buf.WriteRune(r)
			}
		}
	}
	if buf.Len() > 0 {
		res = append(res, buf.String())
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

func GetLastItem[T any](slice []T) T {
	if len(slice) < 1 {
		return *new(T)
	}
	return slice[len(slice)-1]
}

func SplitByToken(reader io.Reader, token string, keepToken bool) <-chan string {
	if len(token) == 0 {
		log.Fatalln("token should not be empty")
	}
	ch := make(chan string)
	tokenBytes := typesw.StrToBytes(token)
	go func() {
		defer close(ch)
		if reader == nil {
			return
		}

		r := bufio.NewReader(reader)
		var buf bytes.Buffer
		for {
			if buf.Len() >= len(tokenBytes) && bytes.Equal(buf.Bytes()[buf.Len()-len(tokenBytes):], tokenBytes) {
				str := buf.String()
				if !keepToken {
					str = str[:len(str)-len(tokenBytes)]
				}
				if len(str) > 0 {
					ch <- str
				}
				buf.Reset()
			}
			d, err := r.ReadBytes(token[0])
			if len(d) == 0 {
				break
			}
			if err != nil && err != io.EOF {
				// log.Fatalln(err)
				break
			}
			end := (err == io.EOF)
			// read next len(tokenBytes)-1
			// n, err := r.Read(rest)
			buf.Write(d)

			rest, err := r.Peek(len(tokenBytes) - 1)
			if len(rest) < len(tokenBytes)-1 || err == io.EOF {
				end = true
			}
			if bytes.Equal(rest, tokenBytes[1:]) {
				buf.Write(rest)
				// fmt.Println("==>")
				// fmt.Println(buf.String())
				// fmt.Println("<==")

				r.Read(rest)
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
