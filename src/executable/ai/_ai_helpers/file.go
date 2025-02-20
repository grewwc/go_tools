package _ai_helpers

import (
	"bytes"
	"strings"

	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/utilsW"
)

type FileParser struct {
	nonTextfiles []string
	textFiles    []string
}

func NewParser(content string) *FileParser {
	files := stringsW.SplitNoEmptyKeepQuote(content, ',')
	ret := FileParser{}
	for _, file := range files {
		file = strings.TrimSpace(file)
		file = utilsW.ExpandUser(file)
		if utilsW.IsTextFile(file) {
			ret.textFiles = append(ret.textFiles, file)
		} else {
			ret.nonTextfiles = append(ret.nonTextfiles, file)
		}
	}
	return &ret
}

func (c *FileParser) TextFiles() []string {
	return c.textFiles
}

func (c *FileParser) NonTextFiles() []string {
	return c.nonTextfiles
}

func (c *FileParser) TextFileContents() string {
	var ret bytes.Buffer
	for _, file := range c.textFiles {
		ret.WriteString(utilsW.ReadString(file))
		ret.WriteRune('\n')
	}
	return ret.String()
}
