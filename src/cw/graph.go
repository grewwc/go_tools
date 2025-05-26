package cw

import (
	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/typesw"
)

type UndirectedGraph[T any] struct {
	v          *Map[T, *Set]
	markdedMap *Map[T, *Set]
	marked     *Set
	groupId    *Map[T, int]

	edgeTo *Map[T, *Map[T, T]]

	cmp typesw.CompareFunc[T]

	cnt int
}

func NewUndirectedGraph[T any](cmp typesw.CompareFunc[T]) *UndirectedGraph[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	return &UndirectedGraph[T]{
		v:          NewMap[T, *Set](),
		markdedMap: NewMap[T, *Set](),
		marked:     NewSet(),

		edgeTo:  NewMap[T, *Map[T, T]](),
		groupId: NewMap[T, int](),
		cmp:     cmp,
	}
}

func (g *UndirectedGraph[T]) AddNode(v T) bool {
	return g.v.PutIfAbsent(v, NewSet())
}

func (g *UndirectedGraph[T]) AddEdge(u, v T) bool {
	if g.Connected(u, v) {
		return false
	}
	g.AddNode(u)
	g.AddNode(v)
	s := g.v.GetOrDefault(u, NewSet())
	s.Add(v)
	g.v.PutIfAbsent(u, s)

	s = g.v.GetOrDefault(v, NewSet())
	s.Add(u)
	return true
}

func (g *UndirectedGraph[T]) Adj(u T) *Set {
	return g.v.GetOrDefault(u, NewSet())
}

func (g *UndirectedGraph[T]) Connected(u, v T) bool {
	marked := g.markdedMap.GetOrDefault(u, nil)
	return marked != nil && marked.Contains(u)
}

func (g *UndirectedGraph[T]) Mark() {
	g.clearState()
	for v := range g.v.Iterate() {
		if !g.marked.Contains(v) {
			g.cnt++
		}
		g.bfsMark(v)
	}
}

func (g *UndirectedGraph[T]) NumGroups() int {
	return g.cnt
}

// Group returns the group id.
// returns -1 if u not existed in the graph.
func (g *UndirectedGraph[T]) Group(u T) int {
	return g.groupId.GetOrDefault(u, -1)
}

func (g *UndirectedGraph[T]) Path(from, to T) []T {
	if !g.Connected(from, to) {
		return nil
	}
	return g.path(from, to)
}

func (g *UndirectedGraph[T]) path(from, to T) []T {
	var res []T
	edgeTo := g.edgeTo.GetOrDefault(from, NewMap[T, T]())
	for x := to; g.cmp(x, from) != 0; x = edgeTo.Get(x) {
		res = append(res, x)
	}
	res = append(res, from)
	algow.Reverse(res)
	return res
}

func (g *UndirectedGraph[T]) bfsMark(u T) {
	marked := g.markdedMap.GetOrDefault(u, NewSet())
	edgeTo := g.edgeTo.GetOrDefault(u, NewMap[T, T]())
	q := NewQueue()
	q.Enqueue(u)
	for !q.Empty() {
		curr := q.Dequeue().(T)
		marked.Add(curr)
		g.marked.Add(curr)
		g.groupId.PutIfAbsent(curr, g.cnt)
		for adj := range g.Adj(curr).Iterate() {
			if !marked.Contains(adj) {
				q.Enqueue(adj)
				edgeTo.Put(adj.(T), curr)
			}
		}
	}
	g.markdedMap.PutIfAbsent(u, marked)
	g.edgeTo.PutIfAbsent(u, edgeTo)
}

func (g *UndirectedGraph[T]) clearState() {
	g.markdedMap.Clear()
	g.edgeTo.Clear()
	g.marked.Clear()
	g.groupId.Clear()
	g.cnt = 0
}
