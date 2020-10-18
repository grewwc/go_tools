package terminalW

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/grewwc/go_tools/src/stringsW"
)

/*******************************************

very similar to go default "flag" packge

difference: optional arguments can be put after positional arguments

don't support multiple value (e.g. -arg 1 2)

********************************************/

type ParsedResults struct {
	Optional   map[string]string
	Positional []string
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

// ParseArgsCmd is more powerful than golang default argparser
func ParseArgs(cmd string, boolOptionals ...string) *ParsedResults {
	firstBoolArg := ""
	for _, boolOptional := range boolOptionals {
		boolOptional = strings.ReplaceAll(boolOptional, "-", "")
		cmdNew := stringsW.Move2EndAll(cmd, fmt.Sprintf("\x00-%s", boolOptional))
		if firstBoolArg == "" && cmdNew != cmd {
			firstBoolArg = boolOptional
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
	return ParseArgs(cmd, boolOptionals...)
}
