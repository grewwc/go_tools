package terminalW

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
)

type ParsedResults struct {
	Optional map[string]string
	// Positional *containerW.Set
	Positional *containerW.OrderedSet
}

func (r ParsedResults) GetFlagVal(flagName string) (string, error) {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	if val, exists := r.Optional[flagName]; exists {
		return val, nil
	}
	return "", errors.New("not exist")
}

func (r ParsedResults) GetFlags() *containerW.OrderedSet {
	res := containerW.NewOrderedSet()
	for k := range r.Optional {
		res.Add(k)
	}
	return res
}

func (r ParsedResults) GetBooleanArgs() *containerW.OrderedSet {
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
func (r ParsedResults) ContainsFlag(flagName string) bool {
	flagName = stringsW.StripPrefix(flagName, "-")
	if len(flagName) > 1 {
		return r.ContainsFlagStrict(flagName)
	}

	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	if _, exists := r.Optional[flagName]; exists {
		return true
	}
	// fmt.Println(r.Optional)

	flagName = stringsW.StripPrefix(flagName, "-")
	s2 := containerW.FromString(flagName)
	for k := range r.Optional {
		s1 := containerW.FromString(k)
		if !s2.MutualExclude(*s1) {
			return true
		}
	}
	return false
}

// ContainsFlagStrict checks if an optional flag is set
// "main.exe -force" ==> [ContainsFlag("-f") == false, ContainsFlag("-force") == true]
func (r ParsedResults) ContainsFlagStrict(flagName string) bool {
	if flagName[0] != '-' {
		flagName = "-" + flagName
	}
	if _, exists := r.Optional[flagName]; exists {
		return true
	}

	return false
}

// GetNumArgs return -1 to signal "there is NO num args (e.g.: -10)"
// if exists, return the LARGEST value
func (r ParsedResults) GetNumArgs() int {
	res := -1
	p := regexp.MustCompile("-(\\d+)")

	for k := range r.Optional {
		if !p.MatchString(k) {
			continue
		}
		k = strings.TrimLeft(k, "-")
		kInt, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		if res < int(kInt) {
			res = int(kInt)
		}
	}
	return res
}

type sortByLen []string

func (slice sortByLen) Len() int {
	return len(slice)
}

func (slice sortByLen) Less(i, j int) bool {
	return len(slice[i]) > len(slice[j])
}

func (slice sortByLen) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func classifyArguments(cmd string, endIdx int) (*containerW.OrderedSet, []string, []string) {
	const (
		positionalMode = iota
		optionalKeyMode
		optionalValMode
		spaceMode
	)
	mode := spaceMode
	var positionals = containerW.NewOrderedSet()
	var keys []string
	var vals []string
	var pBuf bytes.Buffer
	var kBuf bytes.Buffer
	var vBuf bytes.Buffer

	for _, ch := range cmd[:endIdx] {
		switch mode {
		case spaceMode:
			if ch == '\x00' {
				continue
			}
			if ch == '-' {
				mode = optionalKeyMode
				kBuf.WriteRune(ch)
			} else {
				pBuf.WriteRune(ch)
				mode = positionalMode
			}

		case positionalMode:
			if ch == '\x00' {
				mode = spaceMode
				// positionals = append(positionals, pBuf.String())
				positionals.Add(pBuf.String())
				pBuf.Reset()
				continue
			}
			pBuf.WriteRune(ch)

		case optionalKeyMode:
			if ch == '\x00' {
				mode = optionalValMode
				keys = append(keys, kBuf.String())
				kBuf.Reset()
				continue
			}
			kBuf.WriteRune(ch)

		case optionalValMode:
			if ch == '\x00' {
				mode = spaceMode
				vals = append(vals, vBuf.String())
				vBuf.Reset()
				continue
			}
			vBuf.WriteRune(ch)
		}
	}
	rests := stringsW.SplitNoEmpty(cmd[endIdx:], "\x00")
	keys = append(keys, rests...)
	return positionals, keys, vals
}

func parseArgs(cmd string, boolOptionals ...string) *ParsedResults {
	firstBoolArg := ""
	sort.Sort(sortByLen(boolOptionals))
	// fmt.Println("after sort", boolOptionals)
	moved := containerW.NewTrie()

	for _, boolOptional := range boolOptionals {
		boolOptional = strings.ReplaceAll(boolOptional, "-", "")
		// fmt.Println("here", boolOptional, cmd, "moved", moved)
		if moved.StartsWith(boolOptional) {
			// remove boolOptional in "cmd"
			// cmd = strings.ReplaceAll(cmd, fmt.Sprintf("\x00-%s", boolOptional), "")
			continue
		}

		cmdNew := stringsW.Move2EndAll(cmd, fmt.Sprintf("\x00-%s", boolOptional))
		if firstBoolArg == "" && cmdNew != cmd {
			firstBoolArg = boolOptional
		}

		if cmdNew != cmd {
			moved.Insert(boolOptional)
		}
		cmd = cmdNew
	}

	idx := strings.Index(cmd, fmt.Sprintf("\x00-%s", firstBoolArg))
	// fmt.Println("index", idx, "firstboolarg", fmt.Sprintf("\x00-%s", firstBoolArg), cmd)

	if idx == -1 || firstBoolArg == "" {
		idx = len(cmd)
	}
	var res ParsedResults
	// fmt.Println("final", cmd, idx)
	allPositionals, keys, vals := classifyArguments(cmd, idx)
	res.Positional = allPositionals

	res.Optional = make(map[string]string)
	// fmt.Println("keys", keys)
	// fmt.Println("vals", vals)
	for i := range keys {
		key := keys[i]
		if i >= len(vals) {
			res.Optional[key] = ""
		} else {
			res.Optional[key] = vals[i]
		}
	}
	return &res
}

func ParseArgsCmd(boolOptionals ...string) *ParsedResults {
	if len(os.Args) <= 1 {
		return nil
	}
	// fmt.Println("prev", boolOptionals)
	boolOptionals = constructBoolOptional(boolOptionals...)
	// fmt.Println("after", boolOptionals)

	cmd := strings.Join(os.Args[1:], "\x00")
	cmd = "\x00" + cmd + "\x00"
	// move -number to end
	p := regexp.MustCompile("-(\\d+)")
	for _, match := range p.FindAllStringSubmatch(cmd, -1) {
		submatch := match[1]
		boolOptionals = append(boolOptionals, submatch)
	}

	return parseArgs(cmd, boolOptionals...)
}

// ParseArgs takes command line as argument, not from terminal directly
// cmd contains the Programs itself
func ParseArgs(cmd string, boolOptionals ...string) *ParsedResults {
	cmdSlice := stringsW.SplitNoEmptyKeepQuote(cmd, ' ')
	if len(cmdSlice) <= 1 {
		return nil
	}
	// fmt.Println("prev", boolOptionals)
	boolOptionals = constructBoolOptional(boolOptionals...)
	// fmt.Println("after", boolOptionals)
	cmd = strings.Join(cmdSlice[1:], "\x00")
	cmd = "\x00" + cmd + "\x00"

	// move -number to end
	p := regexp.MustCompile("-(\\d+)")
	for _, match := range p.FindAllStringSubmatch(cmd, -1) {
		submatch := match[1]
		boolOptionals = append(boolOptionals, submatch)
	}

	return parseArgs(cmd, boolOptionals...)
}

func hasCommon(l1, l2 []int) bool {
	for _, v1 := range l1 {
		for _, v2 := range l2 {
			if v1 == v2 {
				return true
			}
		}
	}
	return false
}

func constructString(boolOptionals []string, indices []int) string {
	res := ""
	for _, idx := range indices {
		res += boolOptionals[idx]
	}
	return res
}

func constructBoolOptional(boolOptionals ...string) []string {
	l := len(boolOptionals)
	if l < 1 {
		return []string{}
	}

	res := containerW.NewSet()
	l = len(boolOptionals)
	m := make(map[int][][]int)
	m[1] = make([][]int, l)
	for i := 0; i < l; i++ {
		m[1][i] = []int{i}
	}

	for curLen := 2; curLen <= l; curLen++ {
		// count the total size
		cnt := 0
		for i := 1; i < curLen; i++ {
			cnt += len(m[i])
		}
		m[curLen] = make([][]int, 0, cnt*(cnt-1))

		for l1 := 1; l1 < curLen; l1++ {
			l2 := curLen - l1
			if l2 < l1 {
				break
			}
			s1 := m[l1]
			s2 := m[l2]
			for i, ss1 := range s1 {
				for j, ss2 := range s2 {
					if i == j || hasCommon(ss1, ss2) {
						continue
					}
					s12 := append(ss1, ss2...)
					s21 := append(ss2, ss1...)
					str12 := constructString(boolOptionals, s12)
					str21 := constructString(boolOptionals, s21)
					if !res.Contains(str12) {
						m[curLen] = append(m[curLen], s12)
						res.Add(str12)
					}
					if !res.Contains(str21) {
						m[curLen] = append(m[curLen], s21)
						res.Add(str21)
					}
				}
			}
		}
	}
	// fmt.Println(m)
	for _, option := range boolOptionals {
		res.Add(option)
	}
	// fmt.Println(res)
	return res.ToStringSlice()
}
