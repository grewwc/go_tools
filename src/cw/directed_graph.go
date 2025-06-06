package cw

import (
	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/typesw"
)

type DirectedGraph[T any] struct {
	nodes      *Map[T, *Set]
	markdedMap *Map[T, *Set]
	marked     *Set
	edgeTo     *Map[T, T]
	hasCycle   bool

	onStack *OrderedSet
	cycle   []T

	groupId      *Map[T, int]
	reversePost  *Stack
	reverseGraph *DirectedGraph[T]

	cmp        typesw.CompareFunc[T]
	needRemark bool

	componentCnt int
	edgeCnt      int
}

func NewDirectedGraph[T any](cmp typesw.CompareFunc[T]) *DirectedGraph[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	return &DirectedGraph[T]{
		nodes:      NewMap[T, *Set](),
		markdedMap: NewMap[T, *Set](),
		marked:     NewSet(),
		edgeTo:     NewMap[T, T](),
		groupId:    NewMap[T, int](),

		onStack: NewOrderedSet(),

		reversePost: NewStack(8),

		cmp:        cmp,
		needRemark: true,
	}
}

func (g *DirectedGraph[T]) AddNode(u T) bool {
	res := g.nodes.PutIfAbsent(u, NewSet())
	if res {
		g.needRemark = true
	}
	return res
}

func (g *DirectedGraph[T]) DeleteNode(u T) bool {
	if !g.nodes.Contains(u) {
		return false
	}
	g.needRemark = true
	for adj := range g.Adj(u).Iter().Iterate() {
		g.DeleteEdge(u, adj.(T))
	}
	g.nodes.Delete(u)
	return true
}

func (g *DirectedGraph[T]) AddEdge(u, v T) bool {
	g.AddNode(u)
	g.AddNode(v)
	adj := g.nodes.Get(u)
	if adj.Contains(v) {
		return false
	}
	adj.Add(v)
	g.needRemark = true
	g.edgeCnt++
	return true
}

func (g *DirectedGraph[T]) DeleteEdge(u, v T) bool {
	if !g.nodes.Contains(u) || !g.nodes.Contains(v) {
		return false
	}
	g.nodes.Get(u).Delete(v)
	g.edgeCnt--
	g.needRemark = true
	return true
}

func (g *DirectedGraph[T]) Adj(u T) *Set {
	return g.nodes.GetOrDefault(u, NewSet())
}

func (g *DirectedGraph[T]) Mark() {
	g.mark(true, true)
	g.needRemark = false
}

func (g *DirectedGraph[T]) Degree(u T, in bool) int {
	if in {
		g.mark(false, true)
		return g.reverseGraph.Degree(u, false)
	} else {
		return g.nodes.GetOrDefault(u, NewSet()).Size()
	}
}

func (g *DirectedGraph[T]) Reachable(u, v T) bool {
	g.mark(false, false)
	marked := g.markdedMap.GetOrDefault(u, NewSet())
	return marked.Contains(v)
}

func (g *DirectedGraph[T]) StronglyConnected(u, v T) bool {
	return g.groupId.Contains(u) && g.groupId.Contains(v) && g.groupId.Get(u) == g.groupId.Get(v)
}

func (g *DirectedGraph[T]) Sorted() typesw.IterableT[T] {
	g.mark(false, true)
	if g.hasCycle {
		return nil
	}
	return typesw.ToIterable[T](g.reversePost.Iter())
}

func (g *DirectedGraph[T]) Nodes() []T {
	return g.nodes.Keys()
}

func (g *DirectedGraph[T]) Reverse() *DirectedGraph[T] {
	res := g.reverse()
	res.Mark()
	return res
}

func (g *DirectedGraph[T]) NumStrongComponents() int {
	return g.componentCnt
}

func (g *DirectedGraph[T]) NumNodes() int {
	return g.nodes.Size()
}

func (g *DirectedGraph[T]) NumEdges() int {
	return g.edgeCnt
}

func (g *DirectedGraph[T]) StrongComponents() [][]T {
	res := make([][]T, g.NumStrongComponents())
	for t := range g.groupId.IterEntry().Iterate() {
		id := t.Val()
		v := t.Key()
		res[id] = append(res[id], v)
	}
	return res
}

func (g *DirectedGraph[T]) dfsMark(root, u T) {
	marked := g.markdedMap.GetOrDefault(root, NewSet())
	g.markdedMap.PutIfAbsent(root, marked)
	needPush := true
	if g.marked.Contains(u) {
		needPush = false
	}
	g.onStack.Add(u)
	marked.Add(u)
	g.marked.Add(u)
	g.groupId.Put(u, g.componentCnt)
	for adj := range g.Adj(u).Iter().Iterate() {
		if g.hasCycle {
			break
		}
		w := adj.(T)
		if !marked.Contains(adj) {
			g.edgeTo.Put(w, u)
			g.dfsMark(root, w)
		} else if g.onStack.Contains(u) && len(g.cycle) == 0 {
			g.hasCycle = true
			s := NewStack(g.onStack.Size())
			for node := u; g.cmp(node, w) != 0; node = g.edgeTo.Get(node) {
				s.Push(node)
			}
			s.Push(adj)
			s.Push(u)
			for val := range s.Iter().Iterate() {
				g.cycle = append(g.cycle, val.(T))
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
		for adj := range g.Adj(curr).Iter().Iterate() {
			if !marked.Contains(adj) {
				q.Enqueue(adj)
				g.edgeTo.Put(adj.(T), curr)
			}
		}
	}
}

func (g *DirectedGraph[T]) Path(from, to T) []T {
	g.mark(true, false)
	if !g.Reachable(from, to) {
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
	g.mark(false, false)
	return g.hasCycle
}

func (g *DirectedGraph[T]) Cycle() []T {
	g.mark(false, true)
	return g.cycle
}

func (g *DirectedGraph[T]) reverse() *DirectedGraph[T] {
	res := NewDirectedGraph(g.cmp)
	for t := range g.nodes.IterEntry().Iterate() {
		k := t.Key()
		v := t.Val()
		for node := range v.Iter().Iterate() {
			res.AddEdge(node.(T), k)
		}
	}
	return res
}

func (g *DirectedGraph[T]) mark(needPath bool, needReverseMark bool) {
	if !g.needRemark {
		return
	}
	for v := range g.nodes.Iter().Iterate() {
		g.dfsMark(v, v)
		if needPath {
			g.bfsMark(v)
		}
	}
	if needReverseMark {
		g.reverseGraph = g.reverse()
		g.reverseGraph.mark(true, false)
		cp := g.reverseGraph.reverse()
		for v := range g.reversePost.Iter().Iterate() {
			if !cp.marked.Contains(v) {
				cp.dfsMark(v.(T), v.(T))
				cp.componentCnt++
			}
		}
		g.groupId = cp.groupId
		g.componentCnt = cp.componentCnt
	}
}
