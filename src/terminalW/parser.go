package terminalW

import (
	"bytes"
	"fmt"
	"os"
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
	for *cur < len(cmd) && cmd[*cur] != ' ' && cmd[*cur] != '=' {
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
	for *cur < len(cmd) && cmd[*cur] == ' ' {
		(*cur)++
	}

	// consider sentence in "" as a complete part
	inQuote := false
	// fmt.Println("left", cmd[*cur:])
	for *cur < len(cmd) && cmd[*cur] != '=' {
		if cmd[*cur] == ' ' && !inQuote {
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
		defaultMode
	)
	mode := defaultMode
	inQuote := false
	var positional, rest bytes.Buffer
	var allPositionals []string
	var curByte byte
	for cur := 0; cur < len(cmd); cur++ {
		curByte = cmd[cur]
		switch mode {
		case positionalMode:
			if curByte == '"' {
				inQuote = !inQuote
				// fmt.Println("changing in pos")
				continue
			}

			if !inQuote && curByte == ' ' {
				rest.WriteByte(curByte)
				allPositionals = append(allPositionals, positional.String())
				mode = defaultMode
				positional.Reset()
				continue
			}

			positional.WriteByte(curByte)
		case optionalKeyMode:
			if curByte == '"' {
				inQuote = !inQuote
				// rest.WriteByte(curByte) // newly added
				// fmt.Println("changing in option key")
				continue
			}

			if !inQuote && curByte == ' ' {
				mode = optionalValMode
				for cur < len(cmd) && cmd[cur] == ' ' {
					rest.WriteByte(cmd[cur])
					cur++
				}
				cur-- // because cur will be added again in the main loop
				continue
			}

			if curByte == '=' {
				mode = optionalValMode
				rest.WriteByte(curByte)
				continue
			}
			rest.WriteByte(curByte)
		case optionalValMode:
			if curByte == '"' {
				inQuote = !inQuote
				rest.WriteByte(curByte) // newly added
				// fmt.Printf("changing in option val %c%c\n", cmd[cur-2], cmd[cur-1])
				// fmt.Println(inQuote, cur)
				continue
			}

			if !inQuote && curByte == ' ' {
				mode = defaultMode
				rest.WriteByte(curByte)
				continue
			}
			rest.WriteByte(curByte)
		case defaultMode:
			if curByte == '"' {
				inQuote = !inQuote
				// fmt.Println("changing in default")
				continue
			}

			if curByte == ' ' {
				rest.WriteByte(curByte)
			} else if curByte == '-' {
				mode = optionalKeyMode
				rest.WriteByte(curByte)
			} else { // positional mode
				mode = positionalMode
				positional.WriteByte(curByte)
			}
		}
	}
	if positional.String() != "" {
		allPositionals = append(allPositionals, positional.String())
	}
	return allPositionals, rest.String()
}

func Parse(cmd string) *ParsedResults {
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
