package cw

import (
	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/typesw"
)

const (
	panicMsg = "call Mark() first"
)

type DirectedGraph[T any] struct {
	v          *Map[T, *Set]
	markdedMap *Map[T, *Set]
	marked     *Set
	edgeTo     *Map[T, T]
	hasCycle   bool

	onStack *Set
	cycle   []T

	groupId *Map[T, int]

	reversePost *Stack

	cmp        typesw.CompareFunc[T]
	needRemark bool
	cnt        int
}

func NewDirectedGraph[T any](cmp typesw.CompareFunc[T]) *DirectedGraph[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	return &DirectedGraph[T]{
		v:          NewMap[T, *Set](),
		markdedMap: NewMap[T, *Set](),
		marked:     NewSet(),
		edgeTo:     NewMap[T, T](),
		groupId:    NewMap[T, int](),

		onStack: NewSet(),

		reversePost: NewStack(8),

		cmp:        cmp,
		needRemark: true,
	}
}

func (g *DirectedGraph[T]) AddNode(u T) bool {
	res := g.v.PutIfAbsent(u, NewSet())
	if res {
		g.needRemark = true
	}
	return res
}

func (g *DirectedGraph[T]) AddEdge(u, v T) bool {
	g.AddNode(u)
	g.AddNode(v)
	adj := g.v.Get(u)
	if adj.Contains(v) {
		return false
	}
	adj.Add(v)
	g.needRemark = true
	return true
}

func (g *DirectedGraph[T]) Adj(u T) []T {
	adj := g.v.GetOrDefault(u, NewSet())
	res := make([]T, 0, adj.Size())
	for val := range adj.Iterate() {
		res = append(res, val.(T))
	}
	return res
}

func (g *DirectedGraph[T]) Mark() {
	g.mark(true, true)
	g.needRemark = false
}

func (g *DirectedGraph[T]) Connected(u, v T) bool {
	g.checkState()
	marked := g.markdedMap.GetOrDefault(u, NewSet())
	return marked.Contains(v)
}

func (g *DirectedGraph[T]) StronglyConnected(u, v T) bool {
	return g.groupId.Get(u) == g.groupId.Get(v)
}

func (g *DirectedGraph[T]) Sorted() []T {
	g.checkState()
	if g.hasCycle {
		return nil
	}
	res := make([]T, 0, g.reversePost.Size())
	for val := range g.reversePost.Iterate() {
		res = append(res, val.(T))
	}
	return res
}

func (g *DirectedGraph[T]) Nodes() []T {
	return g.v.Keys()
}

func (g *DirectedGraph[T]) Reverse() *DirectedGraph[T] {
	res := g.reverse()
	res.Mark()
	return res
}

func (g *DirectedGraph[T]) dfsMark(root, u T, cycleDetection bool) {
	marked := g.markdedMap.GetOrDefault(root, NewSet())
	g.markdedMap.PutIfAbsent(root, marked)
	needPush := true
	if g.marked.Contains(u) {
		needPush = false
	}
	g.onStack.Add(u)
	marked.Add(u)
	g.marked.Add(u)
	g.groupId.Put(u, g.cnt)
	for _, adj := range g.Adj(u) {
		if cycleDetection && g.hasCycle {
			return
		}
		if !marked.Contains(adj) {
			g.edgeTo.Put(adj, u)
			g.dfsMark(root, adj, cycleDetection)
		} else if g.onStack.Contains(u) {
			g.hasCycle = true
			if cycleDetection {
				s := NewStack(g.onStack.Size())
				for node := u; g.cmp(node, adj) != 0; node = g.edgeTo.Get(node) {
					s.Push(node)
				}
				s.Push(adj)
				s.Push(u)
				for val := range s.Iterate() {
					g.cycle = append(g.cycle, val.(T))
				}
			}
		}
	}
	if needPush {
		g.reversePost.Push(u)
	}
	g.onStack.Delete(u)
}

// bfsMark for path
func (g *DirectedGraph[T]) bfsMark(u T) {
	marked := NewSet()
	g.markdedMap.Put(u, marked)
	g.edgeTo.Clear()
	q := NewQueue()
	q.Enqueue(u)
	for !q.Empty() {
		curr := q.Dequeue().(T)
		marked.Add(curr)
		for _, adj := range g.Adj(curr) {
			if !marked.Contains(adj) {
				q.Enqueue(adj)
				g.edgeTo.Put(adj, curr)
			}
		}
	}
}

func (g *DirectedGraph[T]) Path(from, to T) []T {
	g.checkState()
	if !g.Connected(from, to) {
		return nil
	}
	res := make([]T, 0)
	for v := to; g.cmp(v, from) != 0; v = g.edgeTo.Get(v) {
		res = append(res, v)
	}
	res = append(res, from)
	algow.Reverse(res)
	return res
}

func (g *DirectedGraph[T]) HasCycle() bool {
	g.checkState()
	return g.hasCycle
}

func (g *DirectedGraph[T]) Cycle() []T {
	g.checkState()
	return g.cycle
}

func (g *DirectedGraph[T]) checkState() {
	if g.needRemark {
		panic(panicMsg)
	}
}

func (g *DirectedGraph[T]) reverse() *DirectedGraph[T] {
	res := NewDirectedGraph(g.cmp)
	for t := range g.v.IterateEntry() {
		k := t.Get(0).(T)
		v := t.Get(1).(*Set)
		for node := range v.Iterate() {
			res.AddEdge(node.(T), k)
		}
	}
	return res
}

func (g *DirectedGraph[T]) mark(needPath bool, needReverseMark bool) {
	if !g.needRemark {
		return
	}
	for v := range g.v.Iterate() {
		g.dfsMark(v, v, false)
		if needPath {
			g.bfsMark(v)
		}
	}
	if needReverseMark {
		rg := g.reverse()
		rg.mark(true, false)
		cp := rg.reverse()
		for v := range rg.reversePost.Iterate() {
			if !cp.marked.Contains(v) {
				cp.cnt++
				cp.dfsMark(v.(T), v.(T), true)
			}
		}
		g.groupId = cp.groupId
		g.cycle = cp.cycle
	}
}
