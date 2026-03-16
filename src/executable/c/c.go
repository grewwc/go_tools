package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/grewwc/go_tools/src/strw"
)

const defaultPrec = 16

const (
	piDigits = "3.14159265358979323846264338327950288419716939937510"
	eDigits  = "2.71828182845904523536028747135266249775724709369995"
)

type cliOptions struct {
	expr   string
	file   string
	prec   int
	degree bool
	help   bool
}

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenNumber
	tokenIdentifier
	tokenOperator
	tokenLeftParen
	tokenRightParen
	tokenComma
)

type token struct {
	typ   tokenType
	value string
}

type expressionParser struct {
	tokens []token
	pos    int
	prec   int
	degree bool
}

var builtinFunctions = map[string]struct{}{
	"abs":   {},
	"acos":  {},
	"asin":  {},
	"atan":  {},
	"ceil":  {},
	"cos":   {},
	"exp":   {},
	"floor": {},
	"ln":    {},
	"log":   {},
	"max":   {},
	"min":   {},
	"pow":   {},
	"round": {},
	"sin":   {},
	"sqrt":  {},
	"tan":   {},
}

func parseCLIArgs(args []string) (cliOptions, []string, error) {
	opts := cliOptions{prec: defaultPrec}
	exprArgs := make([]string, 0, len(args))
	afterDoubleDash := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if afterDoubleDash {
			exprArgs = append(exprArgs, arg)
			continue
		}
		if arg == "--" {
			afterDoubleDash = true
			continue
		}

		switch {
		case arg == "-h" || arg == "--help":
			opts.help = true
		case arg == "-deg" || arg == "--deg":
			opts.degree = true
		case arg == "-prec" || arg == "--prec":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("flag %s requires a value", arg)
			}
			i++
			prec, err := parsePrecision(args[i])
			if err != nil {
				return opts, nil, err
			}
			opts.prec = prec
		case strings.HasPrefix(arg, "-prec="):
			prec, err := parsePrecision(strings.TrimPrefix(arg, "-prec="))
			if err != nil {
				return opts, nil, err
			}
			opts.prec = prec
		case strings.HasPrefix(arg, "--prec="):
			prec, err := parsePrecision(strings.TrimPrefix(arg, "--prec="))
			if err != nil {
				return opts, nil, err
			}
			opts.prec = prec
		case arg == "-e" || arg == "--expr":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("flag %s requires a value", arg)
			}
			i++
			opts.expr = args[i]
		case strings.HasPrefix(arg, "-e="):
			opts.expr = strings.TrimPrefix(arg, "-e=")
		case strings.HasPrefix(arg, "--expr="):
			opts.expr = strings.TrimPrefix(arg, "--expr=")
		case arg == "-f" || arg == "--file":
			if i+1 >= len(args) {
				return opts, nil, fmt.Errorf("flag %s requires a value", arg)
			}
			i++
			opts.file = args[i]
		case strings.HasPrefix(arg, "-f="):
			opts.file = strings.TrimPrefix(arg, "-f=")
		case strings.HasPrefix(arg, "--file="):
			opts.file = strings.TrimPrefix(arg, "--file=")
		default:
			exprArgs = append(exprArgs, arg)
		}
	}

	return opts, exprArgs, nil
}

func parsePrecision(raw string) (int, error) {
	prec, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("invalid precision %q", raw)
	}
	if prec < 0 {
		return 0, fmt.Errorf("precision must be >= 0")
	}
	return prec, nil
}

func printHelp(w io.Writer) {
	fmt.Fprintln(w, "usage:")
	fmt.Fprintln(w, "  c [options] <expr>")
	fmt.Fprintln(w, "  c -e <expr>")
	fmt.Fprintln(w, "  c -f <file>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "options:")
	fmt.Fprintln(w, "  -prec, --prec <n>   digits kept after division and float-function output (default: 16)")
	fmt.Fprintln(w, "  -deg, --deg         use degrees for sin/cos/tan and inverse trig")
	fmt.Fprintln(w, "  -e, --expr <expr>   explicit expression input")
	fmt.Fprintln(w, "  -f, --file <file>   read expression from file")
	fmt.Fprintln(w, "  -h, --help          show help")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "functions:")
	fmt.Fprintln(w, "  abs sqrt sin cos tan asin acos atan ln log exp floor ceil round min max pow")
	fmt.Fprintln(w, "constants:")
	fmt.Fprintln(w, "  pi e tau")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "examples:")
	fmt.Fprintln(w, "  c '1+2*3'")
	fmt.Fprintln(w, "  c 1 - 2")
	fmt.Fprintln(w, "  c '(1+2)(3+4)'")
	fmt.Fprintln(w, "  c -deg 'sin(30)+cos(60)'")
	fmt.Fprintln(w, "  echo '2pi + sqrt(81)' | c")
}

func stdinIsTTY() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func resolveExpressionInput(exprFlag, fileFlag string, exprArgs []string, stdin io.Reader, stdinTTY bool) (string, error) {
	if expr := normalizeExpression(exprFlag); expr != "" {
		return expr, nil
	}

	fileFlag = strings.TrimSpace(fileFlag)
	if fileFlag != "" {
		buf, err := os.ReadFile(fileFlag)
		if err != nil {
			return "", err
		}
		expr := normalizeExpression(string(buf))
		if expr == "" {
			return "", fmt.Errorf("empty expression")
		}
		return expr, nil
	}

	if len(exprArgs) > 0 {
		expr := normalizeExpression(strings.Join(exprArgs, " "))
		if expr == "" {
			return "", fmt.Errorf("empty expression")
		}
		return expr, nil
	}

	if stdinTTY {
		return "", fmt.Errorf("missing expression")
	}

	buf, err := io.ReadAll(stdin)
	if err != nil {
		return "", err
	}
	expr := normalizeExpression(string(buf))
	if expr == "" {
		return "", fmt.Errorf("empty expression")
	}
	return expr, nil
}

func normalizeExpression(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, "**", "^")
	input = strings.ReplaceAll(input, "\r", " ")
	input = strings.ReplaceAll(input, "\n", " ")
	input = strings.ReplaceAll(input, "\t", " ")
	return strings.TrimSpace(input)
}

func evaluateExpression(expr string, prec int, degree bool) (string, error) {
	tokens, err := tokenize(expr)
	if err != nil {
		return "", err
	}
	parser := expressionParser{
		tokens: tokens,
		prec:   prec,
		degree: degree,
	}
	return parser.parse()
}

func tokenize(input string) ([]token, error) {
	input = normalizeExpression(input)
	if input == "" {
		return nil, fmt.Errorf("empty expression")
	}

	tokens := make([]token, 0, len(input)+1)
	for i := 0; i < len(input); {
		ch := input[i]
		switch {
		case ch == ' ':
			i++
		case isDigit(ch) || ch == '.':
			start := i
			dotSeen := ch == '.'
			if ch == '.' && (i+1 >= len(input) || !isDigit(input[i+1])) {
				return nil, fmt.Errorf("invalid number near %q", input[start:i+1])
			}
			i++
			for i < len(input) {
				if isDigit(input[i]) {
					i++
					continue
				}
				if input[i] == '.' {
					if dotSeen {
						return nil, fmt.Errorf("invalid number %q", input[start:i+1])
					}
					dotSeen = true
					i++
					continue
				}
				break
			}
			lit := input[start:i]
			if strings.HasPrefix(lit, ".") {
				lit = "0" + lit
			}
			if strings.HasSuffix(lit, ".") {
				lit += "0"
			}
			tokens = append(tokens, token{typ: tokenNumber, value: cleanNumericString(lit)})
		case isLetter(ch) || ch == '_':
			start := i
			i++
			for i < len(input) && (isLetter(input[i]) || isDigit(input[i]) || input[i] == '_') {
				i++
			}
			tokens = append(tokens, token{typ: tokenIdentifier, value: strings.ToLower(input[start:i])})
		case strings.ContainsRune("+-*/%^", rune(ch)):
			tokens = append(tokens, token{typ: tokenOperator, value: string(ch)})
			i++
		case ch == '(':
			tokens = append(tokens, token{typ: tokenLeftParen, value: "("})
			i++
		case ch == ')':
			tokens = append(tokens, token{typ: tokenRightParen, value: ")"})
			i++
		case ch == ',':
			tokens = append(tokens, token{typ: tokenComma, value: ","})
			i++
		default:
			return nil, fmt.Errorf("unexpected character %q", string(ch))
		}
	}
	tokens = append(tokens, token{typ: tokenEOF})
	return insertImplicitMultiplication(tokens), nil
}

func insertImplicitMultiplication(tokens []token) []token {
	if len(tokens) <= 1 {
		return tokens
	}

	res := make([]token, 0, len(tokens)*2)
	for i := 0; i < len(tokens)-1; i++ {
		curr := tokens[i]
		next := tokens[i+1]
		res = append(res, curr)
		if needsImplicitMultiplication(curr, next) {
			res = append(res, token{typ: tokenOperator, value: "*"})
		}
	}
	res = append(res, tokens[len(tokens)-1])
	return res
}

func needsImplicitMultiplication(curr, next token) bool {
	if !endsPrimary(curr) || !startsPrimary(next) {
		return false
	}
	if curr.typ == tokenIdentifier && next.typ == tokenLeftParen && isFunctionName(curr.value) {
		return false
	}
	return true
}

func endsPrimary(tok token) bool {
	return tok.typ == tokenNumber || tok.typ == tokenIdentifier || tok.typ == tokenRightParen
}

func startsPrimary(tok token) bool {
	return tok.typ == tokenNumber || tok.typ == tokenIdentifier || tok.typ == tokenLeftParen
}

func (p *expressionParser) parse() (string, error) {
	result, err := p.parseAddSub()
	if err != nil {
		return "", err
	}
	if p.current().typ != tokenEOF {
		return "", fmt.Errorf("unexpected token %q", p.current().value)
	}
	return cleanNumericString(result), nil
}

func (p *expressionParser) parseAddSub() (string, error) {
	left, err := p.parseMulDiv()
	if err != nil {
		return "", err
	}

	for {
		tok := p.current()
		if tok.typ != tokenOperator || (tok.value != "+" && tok.value != "-") {
			break
		}
		p.advance()
		right, err := p.parseMulDiv()
		if err != nil {
			return "", err
		}
		left, err = applyBinaryOp(left, right, tok.value[0], p.prec)
		if err != nil {
			return "", err
		}
	}
	return left, nil
}

func (p *expressionParser) parseMulDiv() (string, error) {
	left, err := p.parseUnary()
	if err != nil {
		return "", err
	}

	for {
		tok := p.current()
		if tok.typ == tokenOperator && (tok.value == "*" || tok.value == "/" || tok.value == "%") {
			p.advance()
			right, err := p.parseUnary()
			if err != nil {
				return "", err
			}
			left, err = applyBinaryOp(left, right, tok.value[0], p.prec)
			if err != nil {
				return "", err
			}
			continue
		}
		if startsPrimary(tok) {
			right, err := p.parseUnary()
			if err != nil {
				return "", err
			}
			left, err = applyBinaryOp(left, right, '*', p.prec)
			if err != nil {
				return "", err
			}
			continue
		}
		break
	}
	return left, nil
}

func (p *expressionParser) parsePower() (string, error) {
	left, err := p.parsePrimary()
	if err != nil {
		return "", err
	}
	if tok := p.current(); tok.typ == tokenOperator && tok.value == "^" {
		p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return "", err
		}
		return applyBinaryOp(left, right, '^', p.prec)
	}
	return left, nil
}

func (p *expressionParser) parseUnary() (string, error) {
	tok := p.current()
	if tok.typ == tokenOperator {
		switch tok.value {
		case "+":
			p.advance()
			return p.parseUnary()
		case "-":
			p.advance()
			value, err := p.parseUnary()
			if err != nil {
				return "", err
			}
			return cleanNumericString(strw.Neg(value)), nil
		}
	}
	return p.parsePower()
}

func (p *expressionParser) parsePrimary() (string, error) {
	tok := p.current()
	switch tok.typ {
	case tokenNumber:
		p.advance()
		return tok.value, nil
	case tokenIdentifier:
		p.advance()
		if p.current().typ == tokenLeftParen && isFunctionName(tok.value) {
			return p.parseFunctionCall(tok.value)
		}
		if constant, ok := resolveConstant(tok.value, p.prec); ok {
			return constant, nil
		}
		return "", fmt.Errorf("unknown identifier: %s", tok.value)
	case tokenLeftParen:
		p.advance()
		value, err := p.parseAddSub()
		if err != nil {
			return "", err
		}
		if p.current().typ != tokenRightParen {
			return "", fmt.Errorf("missing ')' in expression")
		}
		p.advance()
		return value, nil
	default:
		return "", fmt.Errorf("unexpected token %q", tok.value)
	}
}

func (p *expressionParser) parseFunctionCall(name string) (string, error) {
	p.advance()
	args := make([]string, 0, 2)
	if p.current().typ != tokenRightParen {
		for {
			arg, err := p.parseAddSub()
			if err != nil {
				return "", err
			}
			args = append(args, arg)
			if p.current().typ == tokenComma {
				p.advance()
				continue
			}
			break
		}
	}
	if p.current().typ != tokenRightParen {
		return "", fmt.Errorf("missing ')' after %s", name)
	}
	p.advance()
	return callFunction(name, args, p.prec, p.degree)
}

func (p *expressionParser) current() token {
	if p.pos >= len(p.tokens) {
		return token{typ: tokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *expressionParser) advance() {
	if p.pos < len(p.tokens) {
		p.pos++
	}
}

func applyBinaryOp(left, right string, op byte, prec int) (result string, err error) {
	left = cleanNumericString(left)
	right = cleanNumericString(right)
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("%v", recovered)
		}
	}()

	switch op {
	case '+':
		result = strw.Plus(left, right)
	case '-':
		result = strw.Minus(left, right)
	case '*':
		result = strw.Mul(left, right)
	case '/':
		result = strw.Div(left, right, prec)
	case '%':
		result = strw.Mod(left, right)
	case '^':
		return powStrings(left, right, prec)
	default:
		return "", fmt.Errorf("unsupported operator: %c", op)
	}

	result = cleanNumericString(result)
	if result == "" {
		return "", fmt.Errorf("invalid result for %s %c %s", left, op, right)
	}
	return result, nil
}

func powStrings(base, exponent string, prec int) (string, error) {
	if intExp, err := parseIntegerArg(exponent); err == nil {
		if intExp == 0 {
			return "1", nil
		}
		if intExp < 0 {
			positive, err := powStrings(base, strconv.Itoa(-intExp), prec)
			if err != nil {
				return "", err
			}
			return applyBinaryOp("1", positive, '/', prec)
		}

		result := "1"
		factor := base
		for intExp > 0 {
			if intExp%2 == 1 {
				result = strw.Mul(result, factor)
			}
			intExp /= 2
			if intExp > 0 {
				factor = strw.Mul(factor, factor)
			}
		}
		return cleanNumericString(result), nil
	}

	baseFloat, err := strconv.ParseFloat(base, 64)
	if err != nil {
		return "", fmt.Errorf("pow base must be numeric")
	}
	exponentFloat, err := strconv.ParseFloat(exponent, 64)
	if err != nil {
		return "", fmt.Errorf("pow exponent must be numeric")
	}
	return formatFloatValue(math.Pow(baseFloat, exponentFloat), max(defaultPrec, prec))
}

func callFunction(name string, args []string, prec int, degree bool) (string, error) {
	name = strings.ToLower(name)
	switch name {
	case "abs":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		return absString(args[0]), nil
	case "sqrt":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		if value < 0 {
			return "", fmt.Errorf("sqrt requires a non-negative argument")
		}
		return formatFloatValue(math.Sqrt(value), max(defaultPrec, prec))
	case "sin", "cos", "tan":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		if degree {
			value = value * math.Pi / 180.0
		}
		var result float64
		switch name {
		case "sin":
			result = math.Sin(value)
		case "cos":
			result = math.Cos(value)
		default:
			result = math.Tan(value)
		}
		return formatFloatValue(result, max(defaultPrec, prec))
	case "asin", "acos", "atan":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		var result float64
		switch name {
		case "asin":
			result = math.Asin(value)
		case "acos":
			result = math.Acos(value)
		default:
			result = math.Atan(value)
		}
		if degree {
			result = result * 180.0 / math.Pi
		}
		return formatFloatValue(result, max(defaultPrec, prec))
	case "ln":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		if value <= 0 {
			return "", fmt.Errorf("ln requires a positive argument")
		}
		return formatFloatValue(math.Log(value), max(defaultPrec, prec))
	case "log":
		if len(args) != 1 && len(args) != 2 {
			return "", fmt.Errorf("log expects 1 or 2 arguments")
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		if value <= 0 {
			return "", fmt.Errorf("log requires a positive argument")
		}
		if len(args) == 1 {
			return formatFloatValue(math.Log10(value), max(defaultPrec, prec))
		}
		base, err := parseFloatArg(name, args, 1)
		if err != nil {
			return "", err
		}
		if base <= 0 || base == 1 {
			return "", fmt.Errorf("log base must be positive and not equal to 1")
		}
		return formatFloatValue(math.Log(value)/math.Log(base), max(defaultPrec, prec))
	case "exp":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		return formatFloatValue(math.Exp(value), max(defaultPrec, prec))
	case "floor":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		return formatFloatValue(math.Floor(value), 0)
	case "ceil":
		if err := requireArgCount(name, args, 1); err != nil {
			return "", err
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		return formatFloatValue(math.Ceil(value), 0)
	case "round":
		if len(args) != 1 && len(args) != 2 {
			return "", fmt.Errorf("round expects 1 or 2 arguments")
		}
		value, err := parseFloatArg(name, args, 0)
		if err != nil {
			return "", err
		}
		if len(args) == 1 {
			return formatFloatValue(math.Round(value), 0)
		}
		digits, err := parseIntegerArg(args[1])
		if err != nil {
			return "", fmt.Errorf("round precision must be an integer")
		}
		factor := math.Pow(10, float64(digits))
		result := math.Round(value*factor) / factor
		return formatFloatValue(result, max(prec, max(0, digits)))
	case "min":
		if len(args) < 2 {
			return "", fmt.Errorf("min expects at least 2 arguments")
		}
		best := args[0]
		for _, arg := range args[1:] {
			if compareNumericStrings(arg, best) < 0 {
				best = arg
			}
		}
		return cleanNumericString(best), nil
	case "max":
		if len(args) < 2 {
			return "", fmt.Errorf("max expects at least 2 arguments")
		}
		best := args[0]
		for _, arg := range args[1:] {
			if compareNumericStrings(arg, best) > 0 {
				best = arg
			}
		}
		return cleanNumericString(best), nil
	case "pow":
		if err := requireArgCount(name, args, 2); err != nil {
			return "", err
		}
		return applyBinaryOp(args[0], args[1], '^', prec)
	default:
		return "", fmt.Errorf("unknown function: %s", name)
	}
}

func requireArgCount(name string, args []string, want int) error {
	if len(args) != want {
		return fmt.Errorf("%s expects %d argument(s)", name, want)
	}
	return nil
}

func parseFloatArg(name string, args []string, index int) (float64, error) {
	if index >= len(args) {
		return 0, fmt.Errorf("%s expects more arguments", name)
	}
	value, err := strconv.ParseFloat(cleanNumericString(args[index]), 64)
	if err != nil {
		return 0, fmt.Errorf("%s expects numeric arguments", name)
	}
	return value, nil
}

func parseIntegerArg(raw string) (int, error) {
	normalized := cleanNumericString(raw)
	return strconv.Atoi(normalized)
}

func formatFloatValue(value float64, prec int) (string, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "", fmt.Errorf("result is not a finite number")
	}
	if prec < 0 {
		prec = defaultPrec
	}
	value = snapTinyFloat(value, prec)
	return cleanNumericString(strconv.FormatFloat(value, 'f', prec, 64)), nil
}

func snapTinyFloat(value float64, prec int) float64 {
	eps := math.Pow(10, -float64(max(4, prec+2)))
	if math.Abs(value) < eps {
		return 0
	}
	return value
}

func resolveConstant(name string, prec int) (string, bool) {
	digits := max(defaultPrec, prec)
	switch strings.ToLower(name) {
	case "pi":
		return trimConstantDigits(piDigits, digits), true
	case "e":
		return trimConstantDigits(eDigits, digits), true
	case "tau":
		return trimConstantDigits(strw.Mul("2", piDigits), digits), true
	default:
		return "", false
	}
}

func trimConstantDigits(value string, digits int) string {
	idx := strings.IndexByte(value, '.')
	if idx < 0 {
		return value
	}
	end := idx + digits + 1
	if end >= len(value) {
		return value
	}
	return cleanNumericString(value[:end])
}

func compareNumericStrings(left, right string) int {
	diff := cleanNumericString(strw.Minus(left, right))
	switch {
	case diff == "0":
		return 0
	case strings.HasPrefix(diff, "-"):
		return -1
	default:
		return 1
	}
}

func absString(value string) string {
	value = cleanNumericString(value)
	if strings.HasPrefix(value, "-") {
		return value[1:]
	}
	return value
}

func cleanNumericString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}
	if strings.HasPrefix(value, "+") {
		value = value[1:]
	}
	if strings.HasPrefix(value, "-.") {
		value = "-0" + value[1:]
	}
	if strings.HasPrefix(value, ".") {
		value = "0" + value
	}

	sign := ""
	if strings.HasPrefix(value, "-") {
		sign = "-"
		value = value[1:]
	}

	if strings.Contains(value, ".") {
		value = strings.TrimRight(value, "0")
		value = strings.TrimRight(value, ".")
	}
	value = trimLeadingZerosUnsigned(value)
	if value == "" || value == "0" {
		return "0"
	}
	if sign != "" {
		return sign + value
	}
	return value
}

func trimLeadingZerosUnsigned(value string) string {
	if value == "" {
		return value
	}
	if idx := strings.IndexByte(value, '.'); idx >= 0 {
		intPart := strings.TrimLeft(value[:idx], "0")
		if intPart == "" {
			intPart = "0"
		}
		fracPart := value[idx+1:]
		if fracPart == "" {
			return intPart
		}
		return intPart + "." + fracPart
	}
	value = strings.TrimLeft(value, "0")
	if value == "" {
		return "0"
	}
	return value
}

func isFunctionName(name string) bool {
	_, ok := builtinFunctions[strings.ToLower(name)]
	return ok
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	opts, exprArgs, err := parseCLIArgs(os.Args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if opts.help {
		printHelp(os.Stdout)
		return
	}

	expr, err := resolveExpressionInput(opts.expr, opts.file, exprArgs, os.Stdin, stdinIsTTY())
	if err != nil {
		fmt.Println(err)
		printHelp(os.Stdout)
		return
	}

	result, err := evaluateExpression(expr, opts.prec, opts.degree)
	if err != nil {
		fmt.Println("invalid expression:", err)
		os.Exit(1)
	}
	fmt.Println(result)
}
