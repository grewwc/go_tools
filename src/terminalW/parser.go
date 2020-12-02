package terminalW

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
)

type ParsedResults struct {
	Optional   map[string]string
	Positional []string
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

func classifyArguments(cmd string, endIdx int) ([]string, []string, []string) {
	const (
		positionalMode = iota
		optionalKeyMode
		optionalValMode
		spaceMode
	)
	mode := spaceMode
	var positionals []string
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
				positionals = append(positionals, pBuf.String())
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

	moved := containerW.NewTrie()

	for _, boolOptional := range boolOptionals {
		boolOptional = strings.ReplaceAll(boolOptional, "-", "")
		if moved.StartsWith(boolOptional) {
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
	cmd := strings.Join(os.Args[1:], "\x00")
	cmd = "\x00" + cmd + "\x00"
	return parseArgs(cmd, boolOptionals...)
}
