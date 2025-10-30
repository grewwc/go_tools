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

func SplitByCutset(str, cutset string) []string {

	set := cw.NewSetT[rune]()
	for _, r := range cutset {
		set.Add(r)
	}

	res := make([]string, 0, 1)
	buf := strings.Builder{}
	for _, r := range str {
		if set.Contains(r) {
			if buf.Len() > 0 {
				res = append(res, buf.String())
				buf.Reset()
			}
		} else {
			buf.WriteRune(r)
		}
	}
	if buf.Len() > 0 {
		res = append(res, buf.String())
	}
	return res
}

// SplitByStrKeepQuotes splits a string by a string separator while preserving quoted content
//
// Parameters:
//
//	str: the string to split
//	sep: the string separator to use for splitting
//	symbols: the quote symbols to preserve (e.g. "\"'")
//	keepSymbol: whether to keep the quote symbols in the result
//
// Returns:
//
//	[]string: a slice of strings split by the separator, with quotes preserved according to keepSymbol parameter
func SplitByStrKeepQuotes(str string, sep string, symbols string, keepSymbol bool) []string {
	// precheck removes separator characters from the quote symbols set
	// to prevent conflicts between separator and quote symbols
	precheck := func(symbols *cw.SetT[rune], sep string) {
		for _, r := range sep {
			symbols.Delete(r)
		}
	}

	var res []string
	sepBytes := typesw.StrToBytes(sep)
	inquote := false
	var buf bytes.Buffer
	var prev byte

	s := cw.NewSetT[rune]()
	for _, r := range symbols {
		s.Add(r)
	}

	// Remove any separator characters from quote symbols
	precheck(s, sep)
	// If no quote symbols remain after precheck, return original string
	if s.Len() == 0 {
		return []string{str}
	}

	// Process each rune in the input string
	for i, r := range str {
		// Track previous character for escape sequence handling
		if i > 0 {
			prev = str[i-1]
		}
		// Toggle quote state when encountering unescaped quote symbol
		if s.Contains(r) && prev != '\\' {
			inquote = !inquote
			// Include quote symbol in output if keepSymbol is true
			if keepSymbol {
				buf.WriteRune(r)
			}
		} else {
			buf.WriteRune(r)
			// Check if buffer ends with separator
			if buf.Len() > len(sep) && bytes.Equal(buf.Bytes()[buf.Len()-len(sep):], sepBytes) {
				// Only split if not inside quotes
				if !inquote {
					// Extract content before separator and add to result
					content := buf.String()[:buf.Len()-len(sep)]
					if content != "" {
						res = append(res, content)
					}
					buf.Reset()
				}
			}
		}
	}
	// Add remaining content to result
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
			var end bool
			if len(d) == 0 {
				end = true
			}
			if err != nil && err != io.EOF {
				// log.Fatalln(err)
				break
			}
			end = end || (err == io.EOF)
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
