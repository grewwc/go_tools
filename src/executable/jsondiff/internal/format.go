package internal

import (
	"fmt"
	"regexp"

	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw"
)

func format(parser *terminalw.Parser) {
	fname := parser.MustGetFlagVal("f")
	var text string
	var formatedJ *utilsw.Json
	var err error
	if fname == "" {
		text = utilsw.ReadClipboardText()
		formatedJ, err = utilsw.NewJsonFromString(text)

	} else {
		formatedJ, err = utilsw.NewJsonFromFile(fname)
	}
	if err != nil {
		panic(err)
	}
	formated := formatedJ.StringWithIndent("", "  ")
	if parser.ContainsFlagStrict("one-line") {
		re := regexp.MustCompile(`\s+`)
		formated = re.ReplaceAllString(formated, "")
	}
	fmt.Println(strw.SubStringQuiet(formated, 0, 1024))
	outputFname := fmt.Sprintf("%s_f.json", utilsw.BaseNoExt(fname))
	if fname == "" {
		outputFname = "_f.json"
	}
	fmt.Printf("write file to %s\n", outputFname)
	utilsw.WriteToFile(outputFname, typesw.StrToBytes(formated))
}

func RegisterFormat(parser *terminalw.Parser) {
	parser.On(func(p *terminalw.Parser) bool {
		return p.ContainsFlagStrict("f")
	}).Do(func() { format(parser) })
}
