package terminalW

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/strW"
)

const (
	sep  = '\x00'
	dash = '\x10'
)

type Parser struct {
	Optional   map[string]string
	Positional *containerW.OrderedSet
	*flag.FlagSet
}

func NewParser() *Parser {
	return &Parser{
		Optional:   make(map[string]string),
		Positional: containerW.NewOrderedSet(),
		FlagSet:    flag.NewFlagSet(os.Args[0], flag.ContinueOnError),
	}
}

func (r *Parser) GetFlagVal(flagName string) (string, error) {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	if val, exists := r.Optional[flagName]; exists {
		return val, nil
	}
	return "", errors.New("not exist")
}

func (r *Parser) MustGetFlagValAsInt(flagName string) int {
	resStr, err := r.GetFlagVal(flagName)
	if err != nil {
		panic(err)
	}
	res, err := strconv.Atoi(resStr)
	if err != nil {
		panic(err)
	}
	return res
}

func (r *Parser) GetIntFlagVal(flagName string) int {
	return r.MustGetFlagValAsInt(flagName)
}

func (r *Parser) GetIntFlagValOrDefault(flagName string, val int) int {
	if r.ContainsFlagStrict(flagName) {
		return r.MustGetFlagValAsInt(flagName)
	}
	return val
}

func (r *Parser) GetPositionalArgs(removeNumArg bool) []string {
	if removeNumArg {
		remove := fmt.Sprintf("-%d", r.GetNumArgs())
		copy := r.Positional.ShallowCopy()
		copy.Delete(remove)
		return copy.ToStringSlice()
	}
	return r.Positional.ToStringSlice()
}

func (r *Parser) Empty() bool {
	return len(r.Optional) == 0 && r.Positional.Empty()
}

func (r *Parser) MustGetFlagValAsInt64(flagName string) (res int64) {
	resStr, err := r.GetFlagVal(flagName)
	if err != nil {
		panic(err)
	}
	res, err = strconv.ParseInt(resStr, 10, 64)
	if err != nil {
		panic(err)
	}
	return
}

func (r *Parser) GetFlagValueDefault(flagName string, defaultVal string) string {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	if val, exists := r.Optional[flagName]; exists {
		return val
	}
	return defaultVal
}

func (r *Parser) SetFlagValue(flagName string, val string) {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	r.Optional[flagName] = val
}

func (r *Parser) RemoveFlagValue(flagName string) {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	delete(r.Optional, flagName)
}

func (r *Parser) GetMultiFlagValDefault(flagNames []string, defaultVal string) string {
	var result string
	var err error
	for _, flagName := range flagNames {
		if result, err = r.GetFlagVal(flagName); err == nil {
			return result
		}
	}
	return defaultVal
}

func (r *Parser) MustGetFlagVal(flagName string) string {
	res, err := r.GetFlagVal(flagName)
	if err != nil {
		panic(err)
	}
	return res
}

func (r *Parser) GetFlags() *containerW.OrderedSet {
	res := containerW.NewOrderedSet()
	for k := range r.Optional {
		res.Add(k)
	}
	return res
}

func (r *Parser) GetBooleanArgs() *containerW.OrderedSet {
	res := containerW.NewOrderedSet()
	for k, v := range r.Optional {
		if v == "" {
			if k[0] == '-' {
				res.Add(strings.TrimPrefix(k, "-"))
			} else {
				res.Add("-" + k)
			}
			res.Add(k)
		}
	}
	return res
}

// ContainsFlag checks if an optional flag is set
// "main.exe -force" ==> [ContainsFlag("-f") == true, ContainsFlag("-force") == true]
func (r *Parser) ContainsFlag(flagName string) bool {
	flagName = strW.StripPrefix(flagName, "-")
	buf := bytes.NewBufferString("")
	for option := range r.Optional {
		buf.WriteString(option)
	}
	return strings.Contains(buf.String(), flagName)
}

// ContainsFlagStrict checks if an optional flag is set
// if startswith "--", then use the full name (including the leading "--")
// "main.exe -force" ==> [ContainsFlag("-f") == false, ContainsFlag("-force") == true]
func (r *Parser) ContainsFlagStrict(flagName string) bool {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	if _, exists := r.Optional[flagName]; exists {
		return true
	}
	return false
}

func (r *Parser) ContainsAnyFlagStrict(flagNames ...string) bool {
	for _, flagName := range flagNames {
		if r.ContainsFlagStrict(flagName) {
			return true
		}
	}
	return false
}

func (r *Parser) ContainsAllFlagStrict(flagNames ...string) bool {
	for _, flagName := range flagNames {
		if !r.ContainsFlagStrict(flagName) {
			return false
		}
	}
	return true
}

// CoExists “-lrt”， args = ["l", "r", "t"]，return true
// args 的顺序无关
func (r *Parser) CoExists(args ...string) bool {
outer:
	for optional := range r.Optional {
		optional = strings.TrimPrefix(optional, "-")
		for _, arg := range args {
			newOptional := strings.Replace(optional, arg, "", 1)
			if newOptional == optional {
				continue outer
			}
			optional = newOptional
		}
		if optional == "" {
			return true
		}
	}
	return false
}

// GetNumArgs return -1 to signal "there is NO num args (e.g.: -10)"
func (r *Parser) GetNumArgs() int {
	res := -1
	p := regexp.MustCompile(`-(\d+)`)

	for ik := range r.Positional.Iterate() {
		k := ik.(string)
		if !p.MatchString(k) {
			continue
		}
		k = strings.TrimLeft(k, "-")
		kInt, err := strconv.ParseInt(k, 10, 64)
		if err == nil {
			return int(kInt)
		}
	}
	return res
}

func canConstructByBoolOptionals(key string, boolOptionals ...string) bool {
	key = strings.TrimPrefix(key, string(dash))
	// fmt.Println("test", key)
	if key == "" {
		return true
	}
	for i, boolOptional := range boolOptionals {
		if strings.HasPrefix(key, boolOptional) {
			try := canConstructByBoolOptionals(key[len(boolOptional):], append(boolOptionals[:i], boolOptionals[i+1:]...)...)
			if try {
				// fmt.Println(true)
				return true
			}
		}
	}
	// fmt.Println(false)
	return false
}

func classifyArguments(cmd string, boolOptionals ...string) (*containerW.OrderedSet, []string, []string, []string) {
	// fmt.Println("here", strings.ReplaceAll(cmd, "sep", "|"))
	const (
		positionalMode = iota
		optionalKeyMode
		optionalValMode
		spaceMode
		boolOptionalMode

		StartMode
	)
	prev := StartMode

	mode := spaceMode
	var positionals = containerW.NewOrderedSet()
	var keys []string
	var boolKeys []string
	var vals []string
	var pBuf bytes.Buffer
	var kBuf bytes.Buffer
	var vBuf bytes.Buffer

	for _, ch := range cmd {
		switch mode {
		case spaceMode:
			if ch == sep {
				continue
			}
			if ch == dash {
				mode = optionalKeyMode
				kBuf.WriteRune(ch)
			} else {
				if prev == boolOptionalMode || prev == StartMode || prev == positionalMode || prev == optionalValMode {
					mode = positionalMode
					pBuf.WriteRune(ch)
				} else {
					mode = optionalValMode
					vBuf.WriteRune(ch)
				}
				prev = spaceMode
			}

		case positionalMode:
			if ch == sep {
				mode = spaceMode
				positionals.Add(pBuf.String())
				pBuf.Reset()
				prev = positionalMode
				continue
			}
			pBuf.WriteRune(ch)

		case optionalKeyMode:
			if ch == sep {
				// add boolOptionals check here
				kStr := kBuf.String()
				if canConstructByBoolOptionals(kStr, boolOptionals...) {
					prev = boolOptionalMode
					boolKeys = append(boolKeys, kStr)
				} else {
					prev = optionalKeyMode
					keys = append(keys, kStr)
				}
				mode = spaceMode
				kBuf.Reset()
				continue
			}
			kBuf.WriteRune(ch)

		case optionalValMode:
			if ch == sep {
				mode = spaceMode
				vals = append(vals, vBuf.String())
				vBuf.Reset()
				prev = optionalValMode
				continue
			}
			vBuf.WriteRune(ch)
		}
	}
	// fmt.Println(positionals.ToStringSlice(), boolKeys, keys, vals)
	return positionals, boolKeys, keys, vals
}

func (r *Parser) parseArgs(cmd string, boolOptionals ...string) {
	// flag.Parse()
	normalizedBoolOptionals := make([]string, len(boolOptionals))
	for i, boolArg := range boolOptionals {
		normalizedBoolOptionals[i] = strings.TrimLeft(boolArg, string(dash))
	}
	r.VisitAll(func(f *flag.Flag) {
		key := fmt.Sprintf("-%s%c", f.Name, sep)
		indices := strW.KmpSearch(cmd, key)
		if len(indices) >= 1 {
			for _, idx := range indices {
				substr := strW.SubStringQuiet(cmd, idx, idx+len(key)-1)
				cmd = strings.ReplaceAll(cmd, substr, fmt.Sprintf("%c%s", dash, f.Name))
			}
		}
	})

	// fmt.Println(boolOptionals)
	allPositionals, boolKeys, keys, vals := classifyArguments(cmd, normalizedBoolOptionals...)
	r.Positional = allPositionals

	r.Optional = make(map[string]string)
	// fmt.Println("keys", keys)
	// fmt.Println("vals", vals)
	for i, key := range keys {
		key = strings.ReplaceAll(key, string(dash), "-")
		if i < len(vals) {
			r.Optional[key] = vals[i]
		} else {
			r.Optional[key] = ""
		}
	}
	for _, key := range boolKeys {
		key = strings.ReplaceAll(key, string(dash), "-")
		r.Optional[key] = ""
	}
}

func (r *Parser) ParseArgsCmd(boolOptionals ...string) {
	if len(os.Args) <= 1 {
		return
	}

	args := make([]string, len(os.Args))
	for i, arg := range os.Args {
		args[i] = fmt.Sprintf("%q", arg)
	}
	cmd := strings.Join(args, " ")
	// fmt.Println("here", cmd)

	re := regexp.MustCompile(`\-\d+`)
	numArgs := re.FindString(cmd)
	if len(numArgs) > 0 {
		numArgs = numArgs[1:]
		cmd = strings.ReplaceAll(cmd, fmt.Sprintf("%q", numArgs), "")

		boolOptionals = append(boolOptionals, numArgs)
	}

	r.ParseArgs(cmd, boolOptionals...)
}

// ParseArgs takes command line as argument, not from terminal directly
// cmd contains the Programs itself
func (r *Parser) ParseArgs(cmd string, boolOptionals ...string) {
	cmd = strW.ReplaceAllInQuoteUnchange(cmd, '=', ' ')
	cmdSlice := strW.SplitNoEmptyKeepQuote(cmd, ' ')
	if len(cmdSlice) <= 1 {
		return
	}
	cmd = strings.Join(cmdSlice[1:], fmt.Sprintf("%c", sep))
	cmd = fmt.Sprintf("%c", sep) + cmd + fmt.Sprintf("%c", sep)
	r.parseArgs(cmd, boolOptionals...)
}
