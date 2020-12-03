package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"github.com/grewwc/go_tools/src/windowsW"
)

func init() {
	windowsW.EnableVirtualTerminal()
}

func main() {
	var numOfLines = 10
	parsedResults := terminalW.ParseArgsCmd()
	if parsedResults == nil {
		return
	}

	filenames := parsedResults.Positional.ToStringSlice()

	if nStr, exists := parsedResults.Optional["-n"]; exists {
		// delete(parsedResults.Optional, "-n")
		if nStr == "" {
			return
		}
		n, err := strconv.ParseInt(nStr, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		numOfLines = int(n)
	}

	n := parsedResults.GetNumArgs()
	if n != -1 {
		numOfLines = n
	}

	for _, filename := range filenames {
		if utilsW.IsDir(filename) {
			continue
		}
		f, err := os.Open(filename)
		if err != nil {
			log.Println(err)
			continue
		}
		f.Seek(-1, io.SeekEnd)
		fmt.Println(color.HiGreenString("=======>\t%s\n", filename))

		count := 0
		var byteBuf = make([]byte, 1, 1)
		var buf = make([]byte, 0)
		var resBuf = bytes.NewBuffer(buf)
		lines := containerW.NewStack(numOfLines)

		for count < numOfLines {
			n, _ := f.Read(byteBuf)
			if n < 1 {
				goto END
			}
			b := byteBuf[0]
			resBuf.WriteByte(b)
			if b == '\n' {
				count++
				resStr := resBuf.String()
				lines.Push(resStr)
				resBuf.Reset()
			}
			f.Seek(-2, io.SeekCurrent)
		}

	END:
		f.Close()
		for !lines.Empty() {
			fmt.Print(utilsW.ReverseString(lines.Pop().(string)))
		}
		fmt.Printf("\n\n")
	}
}
