package cw

import (
	"math"

	"github.com/grewwc/go_tools/src/typesw"
)

type WeightedUndirectedGraph[T any] struct {
	*UndirectedGraph[T]

	nodes *Map[T, *Set]

	distTo *Map[T, float64]
	pq     *IndexHeap[T, float64]
	edgeTo *Map[T, *Edge[T]]
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

		pq:     NewIndexHeap[T, float64](nil),
		edgeTo: NewMap[T, *Edge[T]](),
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

	s = g.nodes.GetOrDefault(v, NewSet())
	edge = newEdge(v, u, weight)
	s.Add(edge)
	g.nodes.PutIfAbsent(v, s)

	return res
}

func (g *WeightedUndirectedGraph[T]) Edges() typesw.IterableT[*Edge[T]] {
	return typesw.FuncToIterable(func() <-chan *Edge[T] {
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
	deleted := e.Delete(v)
	e = g.nodes.Get(v)
	deleted = deleted && e.Delete(u)
	return deleted
}

func (g *WeightedUndirectedGraph[T]) dijkstraRelax(v T) {
	for e := range g.nodes.GetOrDefault(v, NewSet()).Iterate() {
		edge := e.(*Edge[T])
		w := edge.Other(v, g.cmp)
		newDist := g.distTo.GetOrDefault(v, math.MaxFloat64) + edge.weight
		if g.distTo.GetOrDefault(w, math.MaxFloat64) > newDist {
			g.distTo.Put(w, newDist)
			g.pq.Update(w, newDist)
			g.edgeTo.Put(w, edge)
		}
	}
}

func (g *WeightedUndirectedGraph[T]) dijkstra(v T) {
	g.pq = NewIndexHeap[T, float64](nil)
	g.pq.Update(v, 0)
	g.distTo = NewMap[T, float64]()
	g.distTo.Put(v, 0)
	for !g.pq.IsEmpty() {
		curr := g.pq.PopIndex()
		g.dijkstraRelax(curr)
	}
}

func (g *WeightedUndirectedGraph[T]) ShortestPath(from, to T) typesw.IterableT[*Edge[T]] {
	if !g.Connected(from, to) {
		return nil
	}
	g.dijkstra(from)
	// fmt.Println(g.edgeTo)
	// fmt.Println(g.distTo)
	s := NewStack(8)
	for curr := to; g.cmp(curr, from) != 0; {
		edge := g.edgeTo.Get(curr)
		s.Push(edge)
		// res = append(res, edge)
		// fmt.Println(edge, curr, g.edgeTo, edge.Other(curr, g.cmp))
		curr = edge.Other(curr, g.cmp)
	}
	// res = append(res, from)
	return typesw.ToIterable[*Edge[T]](s)
}
