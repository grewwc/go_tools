package internal

import (
	"flag"
	"sync"

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

	actionMap *cw.Map[*ConditionFunc, *ActionList]
	// actions *ActionList

	onceFlag *sync.Once

	*flag.FlagSet
}

type ParserOption func(*Parser)

type ConditionFunc func(*Parser) bool
