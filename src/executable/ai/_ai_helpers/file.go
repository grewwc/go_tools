package _ai_helpers

import (
	"bytes"
	"strings"

	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/utilw"
)

type FileParser struct {
	nonTextfiles []string
	textFiles    []string
}

func NewParser(content string) *FileParser {
	files := strw.SplitNoEmptyKeepQuote(content, ',')
	ret := FileParser{}
	for _, file := range files {
		file = strings.TrimSpace(file)
		file = utilw.ExpandUser(file)
		// if utilw.IsTextFile(file) {
		// ret.textFiles = append(ret.textFiles, file)
		// } else {
		ret.nonTextfiles = append(ret.nonTextfiles, file)
		// }
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
		ret.WriteString(utilw.ReadString(file))
		ret.WriteRune('\n')
	}
	return ret.String()
}
