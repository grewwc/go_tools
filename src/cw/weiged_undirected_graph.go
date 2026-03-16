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

	hasNegtiveCycle bool
	negtiveEdge     *Set
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

		negtiveEdge: NewSet(),
	}
}

func (g *WeightedUndirectedGraph[T]) AddEdge(u, v T, weight float64) bool {
	added := g.UndirectedGraph.AddEdge(u, v)
	su := g.nodes.GetOrDefault(u, NewSet())
	sv := g.nodes.GetOrDefault(v, NewSet())
	g.nodes.PutIfAbsent(u, su)
	g.nodes.PutIfAbsent(v, sv)

	si := su.Intersect(sv)
	if !si.Empty() {
		it := si.Iter()
		ch := it.Iterate()
		s := (<-ch).(*Edge[T])
		it.Stop()
		if s.weight < 0 {
			g.negtiveEdge.Delete(s)
		}
		s.weight = weight
		if s.weight < 0 {
			g.negtiveEdge.Add(s)
		}
		return true
	} else if !added {
		return false
	}

	edge := newEdge(u, v, weight, false)
	su.Add(edge)
	sv.Add(edge)
	if weight < 0 {
		g.negtiveEdge.Add(edge)
	}

	return true
}

func (g *WeightedUndirectedGraph[T]) findUndirectedEdge(u, v T) *Edge[T] {
	for e := range g.nodes.GetOrDefault(u, NewSet()).Iter().Iterate() {
		edge := e.(*Edge[T])
		uv := g.cmp(edge.v1, u) == 0 && g.cmp(edge.v2, v) == 0
		vu := g.cmp(edge.v1, v) == 0 && g.cmp(edge.v2, u) == 0
		if uv || vu {
			return edge
		}
	}
	return nil
}

func (g *WeightedUndirectedGraph[T]) Edges() typesw.IterableT[*Edge[T]] {
	return typesw.FuncToIterable(func() chan *Edge[T] {
		ch := make(chan *Edge[T])
		go func() {
			defer close(ch)
			for tup := range g.nodes.IterEntry().Iterate() {
				for e := range tup.Val().Iter().Iterate() {
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
	edge := g.findUndirectedEdge(u, v)
	if edge == nil {
		return false
	}

	e1 := g.nodes.Get(u)
	e2 := g.nodes.Get(v)
	deleted1 := e1 != nil && e1.Delete(edge)
	deleted2 := deleted1
	if g.cmp(u, v) != 0 {
		deleted2 = e2 != nil && e2.Delete(edge)
	}

	if e1 != nil && e1.Empty() {
		g.nodes.Delete(u)
	}
	if e2 != nil && e2.Empty() {
		g.nodes.Delete(v)
	}
	if edge.weight < 0 {
		g.negtiveEdge.Delete(edge)
	}
	return deleted1 && deleted2
}

func (g *WeightedUndirectedGraph[T]) dijkstraRelax(v T) {
	for e := range g.nodes.GetOrDefault(v, NewSet()).Iter().Iterate() {
		edge := e.(*Edge[T])
		w := edge.Other(v, g.cmp)
		var newDist float64
		if g.distTo.GetOrDefault(v, math.MaxFloat64) == math.MaxFloat64 {
			newDist = math.MaxFloat64
		} else {
			newDist = g.distTo.GetOrDefault(v, math.MaxFloat64) + edge.weight
		}
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

func (g *WeightedUndirectedGraph[T]) bellmanFord(v T) {
	g.distTo = NewMap[T, float64]()
	g.distTo.Put(v, 0)
	for i := 0; i < g.NumNodes()-1; i++ {
		for edge := range g.Edges().Iterate() {
			if g.distTo.GetOrDefault(edge.v1, math.MaxFloat64) != math.MaxFloat64 &&
				g.distTo.GetOrDefault(edge.v2, math.MaxFloat64) > g.distTo.GetOrDefault(edge.v1, math.MaxFloat64)+edge.weight {
				g.distTo.Put(edge.v2, g.distTo.GetOrDefault(edge.v1, math.MaxFloat64)+edge.weight)
				g.edgeTo.Put(edge.v2, edge)
			}
		}
	}
	g.hasNegtiveCycle = false
	for edge := range g.Edges().Iterate() {
		if g.distTo.GetOrDefault(edge.v1, math.MaxFloat64) != math.MaxFloat64 &&
			g.distTo.GetOrDefault(edge.v2, math.MaxFloat64) > g.distTo.GetOrDefault(edge.v1, math.MaxFloat64)+edge.weight {
			g.hasNegtiveCycle = true
			break
		}
	}
}

func (g *WeightedUndirectedGraph[T]) HasNegtiveCycle() bool {
	return g.hasNegtiveCycle
}

func (g *WeightedUndirectedGraph[T]) ShortestPath(from, to T) typesw.IterableT[*Edge[T]] {
	g.Mark()
	if !g.Connected(from, to) {
		return typesw.EmptyIterable[*Edge[T]]()
	}
	if g.negtiveEdge.Empty() {
		g.dijkstra(from)
	} else {
		g.bellmanFord(from)
	}
	// fmt.Println(g.edgeTo)
	// fmt.Println(g.distTo)
	s := NewStack[*Edge[T]]()
	for curr := to; g.cmp(curr, from) != 0; {
		edge := g.edgeTo.Get(curr)
		s.Push(edge)
		// fmt.Println(edge, curr, g.edgeTo, edge.Other(curr, g.cmp))
		curr = edge.Other(curr, g.cmp)
	}
	// res = append(res, from)
	return s.Iter()
}
