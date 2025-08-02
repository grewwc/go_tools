package terminalw

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/typesw"
)

const (
	sep   = '\x00'
	quote = '\x05'
	dash  = '\x01'
)

type Parser struct {
	Optional      *cw.OrderedMapT[string, string] // key is prefix with '-'
	Positional    *cw.LinkedList[string]
	defaultValMap *cw.TreeMap[string, string] // key is prefix with '-'

	groups *cw.OrderedMapT[string, *Parser]

	cmd string

	numArg string

	*flag.FlagSet
}

func NewParser() *Parser {
	return &Parser{
		Optional:      cw.NewOrderedMapT[string, string](),
		Positional:    cw.NewLinkedList[string](),
		defaultValMap: cw.NewTreeMap[string, string](nil),
		FlagSet:       flag.NewFlagSet(os.Args[0], flag.ContinueOnError),

		groups: cw.NewOrderedMapT[string, *Parser](),
	}
}

func (p *Parser) PrintDefaults() {
	for entry := range p.groups.Iter().Iterate() {
		fmt.Println(color.YellowString(entry.Key()))
		subP := entry.Val()
		subP.PrintDefaults()
	}
	p.FlagSet.PrintDefaults()
}

func (p *Parser) AddGroup(groupName string) *Parser {
	sub := NewParser()
	p.groups.Put(groupName, sub)
	return sub
}

func (p *Parser) Groups() typesw.IterableT[*Parser] {
	return typesw.FuncToIterable(func() chan *Parser {
		ch := make(chan *Parser)
		go func() {
			for entry := range p.groups.Iter().Iterate() {
				ch <- entry.Val()
			}
			close(ch)
		}()
		return ch
	})
}

func (p *Parser) GetGroupByName(groupName string) *Parser {
	res := p.groups.GetOrDefault(groupName, nil)
	if res == nil {
		return nil
	}
	return res
}

func (r *Parser) GetFlagVal(flagName string) (string, error) {
	if len(flagName) == 0 {
		return "", errors.New("flagName is empty")
	}
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	if r.Optional.Contains(flagName) {
		return r.Optional.Get(flagName), nil
	}
	return "", fmt.Errorf("GetFlagVal: flagName (%s) not exist", flagName)
}

func (r *Parser) GetCmd() string {
	return r.cmd
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

func (r *Parser) GetPositionalArgs(excludeNumArg bool) []string {
	if excludeNumArg {
		remove := fmt.Sprintf("-%d", r.GetNumArgs())
		r.Positional.Delete(remove, nil)
	}
	return r.Positional.ToStringSlice()
}

func (r *Parser) Empty() bool {
	return r.Optional.Empty() && r.Positional.Empty()
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
	return r.Optional.GetOrDefault(flagName, defaultVal)
}

func (r *Parser) SetFlagValue(flagName string, val string) {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	r.Optional.Put(flagName, val)
}

func (r *Parser) RemoveFlagValue(flagName string) {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	r.Optional.Delete(flagName)
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
		return r.GetDefaultValue(flagName)
	}
	return res
}

func (r *Parser) GetFlags() *cw.OrderedSet {
	res := cw.NewOrderedSet()
	for entry := range r.Optional.Iter().Iterate() {
		res.Add(entry.Key())
	}
	return res
}

func (r *Parser) GetBooleanArgs() *cw.OrderedSet {
	res := cw.NewOrderedSet()
	for entry := range r.Optional.Iter().Iterate() {
		k, v := entry.Key(), entry.Val()
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
	flagName = strw.StripPrefix(flagName, "-")
	buf := bytes.NewBufferString("")
	for entry := range r.Optional.Iter().Iterate() {
		buf.WriteString(entry.Key())
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
	return r.Optional.Contains(flagName)
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
	for entry := range r.Optional.Iter().Iterate() {
		optional := entry.Key()
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
	// fmt.Println("numags", r.numArgs)
	if num, err := strconv.Atoi(r.numArg); err == nil {
		return num
	}
	return -1
}

// GetDefaultValue return default value of key.
// If not found, return empty string
func (r *Parser) GetDefaultValue(key string) string {
	if len(key) == 0 {
		return ""
	}
	if key[0] != '-' {
		key = fmt.Sprintf("-%s", key)
	}
	return r.defaultValMap.GetOrDefault(key, "")
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

func classifyArguments(cmd string, boolOptionals ...string) (*cw.LinkedList[string], []string, []string, []string) {
	// fmt.Println("here", strings.ReplaceAll(cmd, string(sep), "|"))
	const (
		positionalMode = iota
		optionalKeyMode
		optionalValMode
		spaceMode
		boolOptionalMode

		startMode
	)
	prev := startMode

	mode := spaceMode
	var positionals = cw.NewLinkedList[string]()
	var keys []string
	var boolKeys []string
	var vals []string
	var pBuf bytes.Buffer
	var kBuf bytes.Buffer
	var vBuf bytes.Buffer

	for _, ch := range cmd {
		if ch == quote {
			continue
		}
		// fmt.Println("beg", ch, prev, mode)
		switch mode {
		case spaceMode:
			if ch == sep {
				continue
			}
			if ch == dash {
				mode = optionalKeyMode
				kBuf.WriteRune(ch)
			} else {
				if prev == boolOptionalMode || prev == startMode || prev == positionalMode || prev == optionalValMode {
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
				positionals.PushBack(pBuf.String())
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
	// fmt.Println(positionals, boolKeys, keys, vals)
	return positionals, boolKeys, keys, vals
}

func (r *Parser) parseArgs(cmd string, boolOptionals ...string) {
	// flag.Parse()
	normalizedBoolOptionals := make([]string, len(boolOptionals))
	for i, boolArg := range boolOptionals {
		normalizedBoolOptionals[i] = strings.TrimLeft(boolArg, string(dash))
	}
	supportedOptions := cw.NewSet()
	r.VisitAll(func(f *flag.Flag) {
		supportedOptions.Add(fmt.Sprintf("%c%s", dash, f.Name))
		key := fmt.Sprintf("-%s%c%c", f.Name, quote, sep)
		// fmt.Println("==>", f.Name, f.DefValue)
		r.defaultValMap.Put(fmt.Sprintf("%v%s", dash, f.Name), f.DefValue)
		indices := strw.KmpSearch(cmd, key, -1)
		indicesQuote := strw.KmpSearch(cmd, fmt.Sprintf("%c%s%c", quote, key, quote), -1)
		indices = append(indices, indicesQuote...)
		// fmt.Println("=====> ", cmd, []byte(cmd), []byte(key), key, indices)
		if len(indices) >= 1 {
			for _, idx := range indices {
				substr := strw.SubStringQuiet(cmd, idx, idx+len(key)-1)
				cmd = strings.ReplaceAll(cmd, substr, fmt.Sprintf("%c%s", dash, f.Name))
			}
		}

	})

	// fmt.Println([]byte(boolOptionals[0]))
	allPositionals, boolKeys, keys, vals := classifyArguments(cmd, normalizedBoolOptionals...)
	r.Positional = allPositionals

	r.Optional = cw.NewOrderedMapT[string, string]()
	// fmt.Println("keys", keys)
	// fmt.Println("vals", vals)
	// fmt.Println("boolKeys", boolKeys)
	// fmt.Println([]byte(allPositionals.ToStringSlice()[0]))
	for i, key := range keys {
		if !supportedOptions.Contains(key) {
			// fmt.Printf("here |%s|", key)
			if i < len(vals) {
				r.Positional.PushBack(fmt.Sprintf("%s %s", key, vals[i]))
			} else {
				r.Positional.PushBack(key)
			}
			// put the key back to positionals
			r.Optional.Delete(key)
			continue
		}
		key = strings.ReplaceAll(key, string(dash), "-")
		if i < len(vals) {
			r.Optional.Put(key, vals[i])
		} else {
			r.Optional.Put(key, r.defaultValMap.GetOrDefault(key, ""))
		}
	}
	for _, key := range boolKeys {
		key = strings.ReplaceAll(key, string(dash), "-")
		defaultVal := r.defaultValMap.GetOrDefault(key, "")
		if defaultVal == "false" {
			r.Optional.Put(key, "true")
		} else {
			r.Optional.Put(key, "false")
		}
	}
}

func (r *Parser) ParseArgsCmd(boolOptionals ...string) {
	start := 1
	if len(os.Args) > 1 && r.groups.Contains(os.Args[1]) {
		start++
		r = r.GetGroupByName(os.Args[1])
	}
	args := make([]string, len(os.Args)-start)
	for i, arg := range os.Args[start:] {
		args[i] = fmt.Sprintf("%c%s%c", quote, arg, quote)
	}
	cmd := strings.Join(args, string(sep))
	// fmt.Println("here", cmd)

	// re := regexp.MustCompile(`\-\d+`)
	// numArgs := re.FindString(cmd)
	// if len(numArgs) > 0 {
	// 	numArgs = numArgs[1:]
	// 	cmd = strings.ReplaceAll(cmd, fmt.Sprintf("%c%s%c", quote, numArgs, quote), "")

	// 	boolOptionals = append(boolOptionals, numArgs)
	// }

	r.ParseArgs(cmd, boolOptionals...)
}

// ParseArgs takes command line as argument, not from terminal directly
// cmd contains the Programs itself
func (r *Parser) ParseArgs(cmd string, boolOptionals ...string) {
	r.cmd = cmd
	// cmd = strw.ReplaceAllInQuoteUnchange(cmd, '=', ' ')
	cmdSlice := strw.SplitNoEmptyPreserveQuote(cmd, ' ', fmt.Sprintf(`"'%c`, quote), true)
	// if len(cmdSlice) <= 1 {
	// 	return
	// }

	cmd = strings.Join(cmdSlice, string(sep))

	re := regexp.MustCompile(`\-\d+`)
	numArgs := re.FindString(cmd)
	if len(numArgs) > 0 {
		r.numArg = numArgs[1:]
		// fmt.Println("waht", r.numArgs, []byte(r.numArgs))
		cmd = strings.Replace(cmd, numArgs, "", 1)
	}

	cmd = fmt.Sprintf("%c", sep) + cmd + fmt.Sprintf("%c", sep)
	r.parseArgs(cmd, boolOptionals...)
}
