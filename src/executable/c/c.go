package main

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
)

const (
	INIT = iota
	NUMBER
	OPERATOR
	BLANK_SPACE
	DOT
)

var prec int

func processInputStr(input string) string {
	input = strings.ReplaceAll(input, "**", "^")
	input = strings.ReplaceAll(input, "--", "+")
	return input + " "
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func reportErr(msg []byte) {
	panic(fmt.Sprintf("invalid expression: %s", stringsW.BytesToString(msg)))
}

func pow(a, b string) string {
	fristVal, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return ""
	}
	secondVal, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(math.Pow(fristVal, secondVal), 'f', -1, 64)
}

func calcWithOp(first, second string, op byte) string {
	var val string
	switch op {
	case '+':
		val = stringsW.Plus(first, second)
	case '-':
		val = stringsW.Minus(first, second)
	case '*':
		val = stringsW.Mul(first, second)
	case '/':
		// val = div(first, second)
		val = stringsW.Div(first, second, prec)
		if val == "" {
			return ""
		}
	case '^':
		secondVal, err := strconv.Atoi(second)
		if err != nil {
			val = pow(first, second)
		} else {
			val = first
			for i := 1; i < secondVal; i++ {
				val = stringsW.Mul(val, first)
			}
		}
	}
	// fmt.Printf("calc: %s %s %s = %s\n", first, string(op), second, val)
	return val
}

func calc(expr []byte) string {
	state := INIT
	opSt := containerW.NewStack(4)
	numSt := containerW.NewStack(8)
	var buf bytes.Buffer
	for i := 0; i < len(expr); {
		ch := expr[i]
		if ch == ' ' {
			if state == NUMBER {
				if buf.Len() > 0 {
					numSt.Push(buf.String())
					buf.Reset()
				}

			}
			i++
			continue
		}
		if ch == '(' {
			if state == NUMBER {
				reportErr(expr)
				return ""
			}
			idx := bytes.Index(expr[i+1:], []byte{')'})
			if idx < 0 {
				reportErr(expr)
				return ""
			}
			idx += i + 1
			nestedResult := calc(stringsW.StringToBytes(processInputStr(stringsW.BytesToString(expr[i+1 : idx]))))
			if nestedResult == "" {
				return ""
			}
			numSt.Push(nestedResult)
			state = NUMBER
			i = idx + 1
		} else if isDigit(ch) {
			buf.WriteByte(ch)
			state = NUMBER
			i++
		} else if ch == '^' {
			if state == OPERATOR {
				reportErr(expr)
				return ""
			}
			if state == NUMBER {
				if buf.Len() > 0 {
					numSt.Push(buf.String())
					buf.Reset()
				}
			}
			opSt.Push(ch)
			i++
			state = OPERATOR
		} else if ch == '*' || ch == '/' {
			if state == OPERATOR {
				reportErr(expr)
				return ""
			}
			if state == NUMBER {
				if buf.Len() > 0 {
					numSt.Push(buf.String())
					buf.Reset()
				}
				if numSt.Size() >= 2 && opSt.Size() >= 1 {
					next := opSt.Top().(byte)
					if next == '^' || next == '*' || next == '/' {
						second := numSt.Pop().(string)
						first := numSt.Pop().(string)
						val := calcWithOp(first, second, opSt.Pop().(byte))
						if val == "" {
							reportErr(expr)
							return ""
						}
						numSt.Push(val)
					}
				}

			}
			opSt.Push(ch)
			i++
			state = OPERATOR
		} else if ch == '+' || ch == '-' {
			if state == OPERATOR {
				// state = NUMBER
				if ch == '-' {
					buf.WriteRune('-')
				}
				i++
				continue
			}
			if state == NUMBER {
				if buf.Len() > 0 {
					numSt.Push(buf.String())
					buf.Reset()
				}
				if numSt.Size() >= 2 && opSt.Size() >= 1 {
					second := numSt.Pop().(string)
					first := numSt.Pop().(string)
					val := calcWithOp(first, second, opSt.Pop().(byte))
					if val == "" {
						reportErr(expr)
						return ""
					}
					numSt.Push(val)
				}
			}
			opSt.Push(ch)
			i++
			state = NUMBER
		} else if ch == '\n' {
			i++
			continue
		} else if ch == '.' {
			if state == DOT {
				reportErr(expr)
				return ""
			}
			buf.WriteByte(ch)
			i++
			state = DOT
		} else {
			reportErr(expr)
			return ""
		}
	}
	// do calculationg
	// fmt.Println("all", numSt)
	for numSt.Size() >= 2 {
		second := numSt.Pop().(string)
		first := numSt.Pop().(string)
		if opSt.Empty() {
			reportErr(expr)
			return ""
		}
		op := opSt.Pop().(byte)
		val := calcWithOp(first, second, op)
		// fmt.Printf("calc: |%s|, |%s|, |%s|, op: %s\n", first, second, val, string(op))
		numSt.Push(val)
	}
	return numSt.Pop().(string)
}

func test() {
	x := "1-1+1"
	x = "2*2"
	res := calc(stringsW.StringToBytes(processInputStr(x)))
	fmt.Println(res)
}

func main() {
	parser := terminalW.NewParser()
	parser.Int("prec", 16, "division precision. default is: 16")
	parser.ParseArgsCmd()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	if parser.Empty() {
		parser.PrintDefaults()
		fmt.Println("c '1+2'")
		return
	}

	expr := parser.Positional.ToStringSlice()[0]
	prec = parser.GetIntFlagValOrDefault("prec", 16)
	if parser.GetNumArgs() != -1 {
		prec = parser.GetNumArgs()
	}

	res := calc(stringsW.StringToBytes(processInputStr(expr)))
	fmt.Println(res)
	// test()

}
