package internal

import (
	"flag"

	"github.com/grewwc/go_tools/src/cw"
)

type Parser struct {
	Optional      *cw.OrderedMapT[string, string] // key is prefix with '-'
	Positional    *cw.LinkedList[string]
	defaultValMap *cw.TreeMap[string, string] // key is prefix with '-'

	groups *cw.OrderedMapT[string, *Parser]

	cmd string

	enableParseNum bool
	numArg         string

	boolOptionSet *cw.Set

	actions *ActionList

	*flag.FlagSet
}

type ParserOption func(*Parser)

type ConditionFunc func(*Parser) bool
