package cw

import (
	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/typesw"
)

type UndirectedGraph[T any] struct {
	nodes   *Map[T, *Set]
	marked  *Set
	groupId *Map[T, int]

	edgeTo *Map[T, T]

	cmp typesw.CompareFunc[T]

	groupCnt int
	edgeCnt  int

	hasCycle   bool
	needRemark bool
}

func NewUndirectedGraph[T any](cmp typesw.CompareFunc[T]) *UndirectedGraph[T] {
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	return &UndirectedGraph[T]{
		nodes:  NewMap[T, *Set](),
		marked: NewSet(),

		edgeTo:  NewMap[T, T](),
		groupId: NewMap[T, int](),
		cmp:     cmp,

		needRemark: true,
	}
}

func (g *UndirectedGraph[T]) AddNode(v T) bool {
	res := g.nodes.PutIfAbsent(v, NewSet())
	if res {
		g.needRemark = true
	}
	return res
}

func (g *UndirectedGraph[T]) AddEdge(u, v T) bool {
	g.AddNode(u)
	g.AddNode(v)
	if g.nodes.Get(u) != nil && g.nodes.Get(u).Contains(v) {
		return false
	}
	s := g.nodes.GetOrDefault(u, NewSet())
	s.Add(v)
	g.nodes.PutIfAbsent(u, s)

	s = g.nodes.GetOrDefault(v, NewSet())
	s.Add(u)
	g.nodes.PutIfAbsent(v, s)
	g.needRemark = true
	g.edgeCnt++
	return true
}

func (g *UndirectedGraph[T]) DeleteEdge(u, v T) bool {
	if !g.nodes.Contains(u) || !g.nodes.Contains(v) {
		return false
	}
	s := g.nodes.Get(u)
	if !s.Contains(v) {
		return false
	}
	g.needRemark = true
	g.edgeCnt--
	s.Delete(v)
	g.nodes.Get(v).Delete(u)
	return true
}

func (g *UndirectedGraph[T]) DeleteNode(u T) bool {
	if !g.nodes.Contains(u) {
		return false
	}
	g.needRemark = true
	for adj := range g.Adj(u).Iterate() {
		g.DeleteEdge(u, adj.(T))
	}
	g.nodes.Delete(u)
	return true
}

func (g *UndirectedGraph[T]) Adj(u T) *Set {
	return g.nodes.GetOrDefault(u, NewSet())
}

func (g *UndirectedGraph[T]) Nodes() []T {
	return g.nodes.Keys()
}

func (g *UndirectedGraph[T]) NumNodes() int {
	return g.nodes.Size()
}

func (g *UndirectedGraph[T]) NumEdges() int {
	return g.edgeCnt
}

func (g *UndirectedGraph[T]) Degree(u T) int {
	return g.nodes.Get(u).Size()
}

func (g *UndirectedGraph[T]) Connected(u, v T) bool {
	g.Mark()
	return g.groupId.Contains(u) && g.groupId.Contains(v) &&
		g.groupId.Get(u) == g.groupId.Get(v)
}

func (g *UndirectedGraph[T]) Mark() {
	if !g.needRemark {
		return
	}
	for v := range g.nodes.Iterate() {
		if !g.marked.Contains(v) {
			g.groupCnt++
			g.bfsMark(v)
		}
	}
	g.needRemark = false
}

func (g *UndirectedGraph[T]) NumGroups() int {
	g.Mark()
	return g.groupCnt
}

// Group returns the group id.
// returns -1 if u not existed in the graph.
func (g *UndirectedGraph[T]) Group(u T) int {
	g.Mark()
	return g.groupId.GetOrDefault(u, 0) - 1
}

func (g *UndirectedGraph[T]) Groups() [][]T {
	g.Mark()
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
	g.Mark()
	if !g.Connected(from, to) {
		return nil
	}
	return g.path(from, to)
}

func (g *UndirectedGraph[T]) HasCycle() bool {
	g.Mark()
	return g.hasCycle
}

func (g *UndirectedGraph[T]) path(from, to T) []T {
	s := NewOrderedSet()
	// fmt.Println("here", g.edgeTo)
	for x := to; g.cmp(x, from) != 0 && !s.Contains(x); {
		// res = append(res, x)
		s.Add(x)
		x = g.edgeTo.Get(x)
	}
	// res = append(res, from)
	s.Add(from)
	res := make([]T, 0, s.Size())
	for node := range s.Iterate() {
		res = append(res, node.(T))
	}
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
		g.groupId.PutIfAbsent(curr, g.groupCnt)
		for adj := range g.Adj(curr).Iterate() {
			if !g.marked.Contains(adj) {
				q.Enqueue(adj)
				g.edgeTo.Put(adj.(T), curr)
				g.edgeTo.Put(curr, adj.(T))
				parent.Put(adj.(T), curr)
			} else if g.cmp(adj.(T), parent.Get(curr)) != 0 {
				g.hasCycle = true
			}
		}
	}
}
