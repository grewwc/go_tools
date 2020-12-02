package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
)

func init() {
	terminalW.EnableVirtualTerminal()
}

func main() {
	var numOfLines = 10
	parsedResults := terminalW.ParseArgsCmd()
	if parsedResults == nil {
		return
	}

	filenames := parsedResults.Positional

	if nStr, exists := parsedResults.Optional["-n"]; exists {
		if nStr == "" {
			return
		}
		n, err := strconv.ParseInt(nStr, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		numOfLines = int(n)
	}

	for k := range parsedResults.Optional {
		k = strings.TrimLeft(k, "-")
		kInt, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			log.Fatalln(err)
		}
		if numOfLines > int(kInt) {
			numOfLines = int(kInt)
		}
	}

	for _, filename := range filenames {
		f, err := os.Open(filename)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(color.HiGreenString("=======>\t%s\n", filename))
		scanner := bufio.NewScanner(f)
		count := 0
		for scanner.Scan() && count < numOfLines {
			line := scanner.Text()
			count++
			fmt.Printf("\t%s\n", line)
		}

		f.Close()

		fmt.Printf("\n\n")
	}
}
