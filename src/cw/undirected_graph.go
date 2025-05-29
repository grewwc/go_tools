package cw

import (
	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/typesw"
)

type UndirectedGraph[T any] struct {
	v       *Map[T, *Set]
	marked  *Set
	groupId *Map[T, int]

	edgeTo *Map[T, T]

	cmp        typesw.CompareFunc[T]
	cnt        int
	hasCycle   bool
	needRemark bool
}

func NewUndirectedGraph[T any](cmp typesw.CompareFunc[T]) *UndirectedGraph[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	return &UndirectedGraph[T]{
		v:      NewMap[T, *Set](),
		marked: NewSet(),

		edgeTo:  NewMap[T, T](),
		groupId: NewMap[T, int](),
		cmp:     cmp,

		needRemark: true,
	}
}

func (g *UndirectedGraph[T]) AddNode(v T) bool {
	res := g.v.PutIfAbsent(v, NewSet())
	if res {
		g.needRemark = true
	}
	return res
}

func (g *UndirectedGraph[T]) AddEdge(u, v T) bool {
	g.AddNode(u)
	g.AddNode(v)
	s := g.v.GetOrDefault(u, NewSet())
	s.Add(v)
	g.v.PutIfAbsent(u, s)

	s = g.v.GetOrDefault(v, NewSet())
	s.Add(u)
	g.needRemark = true
	return true
}

func (g *UndirectedGraph[T]) Adj(u T) *Set {
	return g.v.GetOrDefault(u, NewSet())
}

func (g *UndirectedGraph[T]) Nodes() []T {
	return g.v.Keys()
}

func (g *UndirectedGraph[T]) Degree(u T) int {
	return g.v.Get(u).Size()
}

func (g *UndirectedGraph[T]) Connected(u, v T) bool {
	g.checkState()
	return g.groupId.Get(u) == g.groupId.Get(v)
}

func (g *UndirectedGraph[T]) Mark() {
	if !g.needRemark {
		return
	}
	for v := range g.v.Iterate() {
		if !g.marked.Contains(v) {
			g.cnt++
		}
		g.bfsMark(v)
	}
	g.needRemark = false
}

func (g *UndirectedGraph[T]) NumGroups() int {
	g.checkState()
	return g.cnt
}

// Group returns the group id.
// returns -1 if u not existed in the graph.
func (g *UndirectedGraph[T]) Group(u T) int {
	g.checkState()
	return g.groupId.GetOrDefault(u, 0) - 1
}

func (g *UndirectedGraph[T]) Groups() [][]T {
	g.checkState()
	m := NewMap[int, []T]()
	for node := range g.groupId.data {
		id := g.Group(node.(T))
		arr := m.GetOrDefault(id, make([]T, 0))
		arr = append(arr, node.(T))
		m.Put(id, arr)
	}
	res := make([][]T, m.Size())
	for i := 0; i < m.Size(); i++ {
		res[i] = append(res[i], m.Get(i)...)
	}
	return res
}

func (g *UndirectedGraph[T]) Path(from, to T) []T {
	g.checkState()
	if !g.Connected(from, to) {
		return nil
	}
	return g.path(from, to)
}

func (g *UndirectedGraph[T]) HasCycle() bool {
	g.checkState()
	return g.hasCycle
}

func (g *UndirectedGraph[T]) path(from, to T) []T {
	var res []T
	for x := to; g.cmp(x, from) != 0; x = g.edgeTo.Get(x) {
		res = append(res, x)
	}
	res = append(res, from)
	algow.Reverse(res)
	return res
}

func (g *UndirectedGraph[T]) bfsMark(u T) {
	// marked := g.markdedMap.GetOrDefault(u, NewSet())
	// g.markdedMap.PutIfAbsent(u, marked)
	parent := NewMap[T, T]()
	q := NewQueue()
	q.Enqueue(u)
	for !q.Empty() {
		curr := q.Dequeue().(T)
		g.marked.Add(curr)
		g.marked.Add(curr)
		g.groupId.PutIfAbsent(curr, g.cnt)
		for adj := range g.Adj(curr).Iterate() {
			if !g.marked.Contains(adj) {
				q.Enqueue(adj)
				g.edgeTo.Put(adj.(T), curr)
				parent.Put(adj.(T), curr)
			} else if g.cmp(adj.(T), parent.Get(curr)) != 0 {
				g.hasCycle = true
			}
		}
	}
}

func (g *UndirectedGraph[T]) checkState() {
	if g.needRemark {
		panic("call Mark() first")
	}
}
