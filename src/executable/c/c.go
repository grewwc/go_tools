package main

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
)

const (
	INIT = iota
	NUMBER
	OPERATOR
	BLANK_SPACE
)

func processInputStr(input string) string {
	input = strings.ReplaceAll(input, "**", "^")
	return input + " "
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func reportErr(msg []byte) {
	panic(fmt.Sprintf("invalid expression: %s", stringsW.BytesToString(msg)))
}

func div(a, b string) string {
	fristVal, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return ""
	}
	secondVal, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return ""
	}
	return strconv.FormatFloat(fristVal/secondVal, 'f', -1, 64)
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
		val = div(first, second)
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
			state = BLANK_SPACE
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
				if numSt.Size() >= 2 && opSt.Size() >= 1 && opSt.Top().(byte) == '^' {
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
			state = OPERATOR
		} else if ch == '+' || ch == '-' {
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
					prevOp := opSt.Top().(byte)
					if prevOp != '+' && prevOp != '-' {
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
			state = NUMBER
		} else if ch == '\n' {
			i++
			continue
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("c '1+2'")
		return
	}
	expr := os.Args[1]
	res := calc(stringsW.StringToBytes(processInputStr(expr)))
	fmt.Println(res)
	// x := "1*4*9^22"
	// res := calc(stringsW.StringToBytes(processInputStr(x)))
	// fmt.Println(res)
}
