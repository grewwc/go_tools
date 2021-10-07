package utils

import "encoding/xml"

type properties struct {
	XMLName    xml.Name `xml:"properties"`
	JavaSource int      `xml:"maven.compiler.source"`
	JavaTarget int      `xml:"maven.compiler.target"`
}

func NewProperties() *properties {
	p := &properties{JavaSource: 8, JavaTarget: 8}
	return p
}
