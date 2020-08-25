package configW

import "regexp"

const (
	comment1     = "//"
	comment2     = "#"
	mapSeparator = "->"
	attrFile = ".attr"
)

var reVar = regexp.MustCompile(`(\$\{(.*?)\})`)
var reAttr = regexp.MustCompile(`\[(.*?)\]`)
