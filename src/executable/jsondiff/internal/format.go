package internal

import (
	"fmt"
	"os"

	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func format(parser *terminalw.Parser) {
	fname := parser.MustGetFlagVal("f")
	var text string
	var formatedJ *utilsw.Json
	var err error
	var f *os.File
	if fname == "" {
		text = utilsw.ReadClipboardText()
		formatedJ, err = utilsw.NewJsonFromString(text)

	} else {
		f, err = os.Open(fname)
		if err != nil {
			panic(err)
		}
		// defer f.Close()
		formatedJ, err = utilsw.NewJsonFromReader(f)
	}
	if err != nil {
		panic(err)
	}
	formated := formatedJ.StringWithIndent("", "  ")
	if len(text) < 1024*8 {
		fmt.Println(formated)
	}
	outputFname := fmt.Sprintf("%s_f.json", fname)
	if outputFname == "" {
		outputFname = "_f.json"
		fmt.Printf("write file to %s\n", outputFname)
	}
	utilsw.WriteToFile(outputFname, typesw.StrToBytes(formated))
}

func RegisterFormat(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return p.ContainsFlagStrict("f")
	}).Do(func() { format(parser) })
}
