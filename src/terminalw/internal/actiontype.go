package internal

import "github.com/grewwc/go_tools/src/cw"

var EmptyAction ActionFunc = func() {}

type ActionList struct {
	actionList         *cw.LinkedList[ActionFunc]
	parentParser       *Parser
	conditionSatisfied bool
}

type ActionFunc func()

func (at *ActionList) Do(f ActionFunc) {
	if !at.conditionSatisfied {
		return
	}
	if f == nil {
		f = EmptyAction
	}
	at.actionList.PushBack(f)
}

func newActionList(p *Parser) *ActionList {
	return &ActionList{
		actionList:   cw.NewLinkedList[ActionFunc](),
		parentParser: p,
	}
}
