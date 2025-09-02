package terminalw

import (
	"github.com/grewwc/go_tools/src/terminalw/internal"
)

type Parser struct {
	*internal.Parser
}

type ParserOption func(*Parser)

func NewParser(options ...ParserOption) *Parser {
	iOptions := make([]internal.ParserOption, len(options))
	r := &Parser{}
	for i, option := range options {
		iOptions[i] = internal.ParserOption(func(p *internal.Parser) {
			option(r)
		})
	}
	r.Parser = internal.NewParser()
	r.WithOptions(iOptions...)
	return r
}

func DisableParserNumber(p *Parser) {
	internal.DisableParserNumber(p.Parser)
}

// On should be called before Parsing, otherwise need to explicitly call Execute
func (r *Parser) On(condition func(p *Parser) bool) *internal.ActionList {
	ff := func(*internal.Parser) bool {
		return condition(r)
	}
	return r.Parser.On(internal.ConditionFunc(ff))
	// return nil
}
