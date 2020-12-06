package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

func processSingle(filename string, numOfLines int) {
	if utilsW.IsDir(filename) {
		return
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Println(err)
		return
	}
	cursor, err := f.Seek(-1, io.SeekEnd)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Println(color.HiGreenString("=======>\t%s\n", filename))

	count := 0
	var byteBuf = make([]byte, 1, 1)
	var buf = make([]byte, 0)
	var resBuf = bytes.NewBuffer(buf)
	lines := containerW.NewStack(numOfLines)
	for count < numOfLines {
		n, err := f.Read(byteBuf)
		cursor += int64(n)
		if err != nil {
			log.Println(err)
			return
		}
		if cursor < 0 {
			goto END
		}
		// fmt.Println(cursor)

		b := byteBuf[0]
		resBuf.WriteByte(b)
		if b == '\n' {
			count++
			resStr := resBuf.String()
			lines.Push(resStr)
			resBuf.Reset()
		}

		f.Seek(-2, io.SeekCurrent)
		cursor -= 2
	}

END:
	f.Close()
	for !lines.Empty() {
		fmt.Print(utilsW.ReverseString(lines.Pop().(string)))
	}
	fmt.Printf("\n\n")
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
		fnameMap := utilsW.LsDirGlob(filename)
		for d, fnames := range fnameMap {
			for _, fname := range fnames {
				fname = filepath.Join(d, fname)
				processSingle(fname, numOfLines)
			}
		}
	}
}
