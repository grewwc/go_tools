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

// find the next 2 values from "cmd"
// ignore "=", " "
func findKV(cmd string, cur *int) (string, string) {
	for *cur < len(cmd) && cmd[*cur] != '-' {
		(*cur)++
	}

	(*cur)++
	var key, val bytes.Buffer
	for *cur < len(cmd) && cmd[*cur] != '\x00' && cmd[*cur] != '=' {
		key.WriteByte(cmd[*cur])
		(*cur)++
	} // key has been parsed

	// go pass '=' sign
	if *cur < len(cmd) && cmd[*cur] == '=' {
		if *cur+1 < len(cmd) && cmd[*cur+1] == '=' {
			fmt.Fprintf(os.Stdout, "argument error: %s\n", key.String())
			os.Exit(2)
		}
		(*cur)++
	}

	// following is for parsing dict value

	// ignore more empty space
	for *cur < len(cmd) && cmd[*cur] == '\x00' {
		(*cur)++
	}

	// consider sentence in "" as a complete part
	inQuote := false
	// fmt.Println("left", cmd[*cur:])
	for *cur < len(cmd) && cmd[*cur] != '=' {
		if cmd[*cur] == '\x00' && !inQuote {
			break
		}
		if cmd[*cur] == '"' {
			if !inQuote {
				inQuote = true
			} else {
				inQuote = false
			}
		} else if cmd[*cur] != '-' || inQuote {
			val.WriteByte(cmd[*cur])
		}
		(*cur)++
	} // val has been parsed
	// fmt.Println("here", key.String(), val.String())
	return key.String(), val.String()
}

// 1 return: all positional arguments
// 2 return: rest command line string
// IMPORTNAT: boolean args needs to put to end  !!!!!!!!
func findAllPositonalArguments(cmd string) ([]string, string) {
	const (
		positionalMode = iota
		optionalKeyMode
		optionalValMode
		spaceMode
	)
	mode := spaceMode
	var positionals []string
	var pBuf bytes.Buffer
	var kvBuf bytes.Buffer

	for _, ch := range cmd {
		switch mode {
		case spaceMode:
			if ch == '\x00' {
				continue
			}
			if ch == '-' {
				mode = optionalKeyMode
				kvBuf.WriteRune(ch)
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
			kvBuf.WriteRune(ch)
			if ch == '\x00' {
				mode = optionalValMode
			}
		case optionalValMode:
			kvBuf.WriteRune(ch)
			if ch == '\x00' {
				mode = spaceMode
			}
		}
	}
	return positionals, kvBuf.String()
}

// ParseArgs is more powerful than golang default argparser
func ParseArgs(boolOptionals ...string) *ParsedResults {
	if len(os.Args) <= 1 {
		return nil
	}
	cmd := strings.Join(os.Args[1:], "\x00")
	cmd = "\x00" + cmd + "\x00"
	for _, boolOptional := range boolOptionals {
		boolOptional = strings.ReplaceAll(boolOptional, "-", "")
		cmd = stringsW.Move2EndAll(cmd, fmt.Sprintf("\x00-%s", boolOptional))
	}
	var cur int
	var k, v string
	var res ParsedResults

	allPositionals, rest := findAllPositonalArguments(cmd)
	res.Positional = allPositionals

	res.Optional = make(map[string]string)
	cmd = rest
	// fmt.Println("heere ", cmd)
	for cur < len(cmd) {
		k, v = findKV(cmd, &cur)
		if k == "" {
			continue
		}
		res.Optional[k] = v
	}
	return &res
}
