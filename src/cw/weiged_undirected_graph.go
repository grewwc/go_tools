package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

type WeightedUndirectedGraph[T any] struct {
	*UndirectedGraph[T]

	nodes *Map[T, *Set]
}

type mst[T any] struct {
	edges []*Edge[T]
}

func (m *mst[T]) Edges() []*Edge[T] {
	return m.edges
}

func (m *mst[T]) TotalWeight() float64 {
	total := 0.0
	for _, edge := range m.edges {
		total += edge.weight
	}
	return total
}

func NewWeightedUndirectedGraph[T any](cmp typesw.CompareFunc[T]) *WeightedUndirectedGraph[T] {
	return &WeightedUndirectedGraph[T]{
		UndirectedGraph: NewUndirectedGraph(cmp),
		nodes:           NewMap[T, *Set](),
	}
}

func (g *WeightedUndirectedGraph[T]) AddEdge(u, v T, weight float64) bool {
	res := g.UndirectedGraph.AddEdge(u, v)
	if !res {
		return false
	}
	s := g.nodes.GetOrDefault(u, NewSet())
	edge := newEdge(u, v, weight)
	s.Add(edge)
	g.nodes.PutIfAbsent(u, s)

	return res
}

func (g *WeightedUndirectedGraph[T]) Edges() typesw.IterableT[*Edge[T]] {
	return typesw.ToIterable(func() <-chan *Edge[T] {
		ch := make(chan *Edge[T])
		go func() {
			defer close(ch)
			for tup := range g.nodes.IterateEntry() {
				for e := range tup.Get(1).(*Set).Iterate() {
					ch <- e.(*Edge[T])
				}
			}
		}()
		return ch
	})
}

func (g *WeightedUndirectedGraph[T]) Mst() *mst[T] {
	edgeCmp := func(e1, e2 *Edge[T]) int {
		return e1.cmp(e2)
	}
	uf := NewUF(g.cmp)
	h := NewHeap(edgeCmp)
	for edge := range g.Edges().Iterate() {
		h.Insert(edge)
	}
	s := make([]*Edge[T], 0, g.NumEdges()-1)
	for !h.IsEmpty() && len(s) < g.NumNodes()-1 {
		edge := h.Pop()
		if uf.IsConnected(edge.v1, edge.v2) {
			continue
		}
		uf.Union(edge.v1, edge.v2)
		s = append(s, edge)
	}
	return &mst[T]{
		edges: s,
	}
}

func (g *WeightedUndirectedGraph[T]) DeleteEdge(u, v T) bool {
	res := g.UndirectedGraph.DeleteEdge(u, v)
	if !res {
		return res
	}
	e := g.nodes.Get(u)
	if e == nil {
		return false
	}
	return e.Delete(v)
}
