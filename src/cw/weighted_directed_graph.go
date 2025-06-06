package cw

import (
	"math"

	"github.com/grewwc/go_tools/src/typesw"
)

type WeightedDirectedGraph[T any] struct {
	*DirectedGraph[T]

	nodes  *Map[T, *Set]
	distTo *Map[T, float64]
	pq     *IndexHeap[T, float64]
	edgeTo *Map[T, *Edge[T]]

	hasNegtiveCycle bool
	negtiveEdge     *Set

	um *Map[T, *Edge[T]]
}

func NewWeightedDirectedGraph[T any](cmp typesw.CompareFunc[T]) *WeightedDirectedGraph[T] {
	return &WeightedDirectedGraph[T]{
		DirectedGraph: NewDirectedGraph(cmp),

		nodes:  NewMap[T, *Set](),
		distTo: NewMap[T, float64](),
		pq:     NewIndexHeap[T, float64](nil),
		edgeTo: NewMap[T, *Edge[T]](),

		negtiveEdge: NewSet(),

		um: NewMap[T, *Edge[T]](),
	}
}

func (g *WeightedDirectedGraph[T]) AddEdge(u, v T, weight float64) bool {
	if !g.DirectedGraph.AddEdge(u, v) {
		return false
	}
	edge := newEdge(u, v, weight, true)
	s := g.nodes.GetOrDefault(u, NewSet())
	g.nodes.PutIfAbsent(u, s)
	s.Add(edge)
	g.um.Put(u, edge)
	if weight < 0 {
		g.negtiveEdge.Add(edge)
	}
	return true
}

func (g *WeightedDirectedGraph[T]) DeleteEdge(u, v T) bool {
	if !g.DirectedGraph.DeleteEdge(u, v) {
		return false
	}
	if !g.nodes.Contains(u) {
		return false
	}
	e := g.um.Get(u)
	g.nodes.Get(u).Delete(e)
	if g.nodes.Get(u).Empty() {
		g.nodes.Delete(u)
		g.um.Delete(u)
	}
	if e.weight < 0 {
		g.negtiveEdge.Delete(e)
	}
	return true
}

func (g *WeightedDirectedGraph[T]) Edges() typesw.IterableT[*Edge[T]] {
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

func (g *WeightedDirectedGraph[T]) dijkstraRelax(v T) {
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

func (g *WeightedDirectedGraph[T]) dijkstra(v T) {
	g.pq = NewIndexHeap[T, float64](nil)
	g.pq.Update(v, 0)
	g.distTo.Clear()
	g.distTo.Put(v, 0)
	for !g.pq.IsEmpty() {
		curr := g.pq.PopIndex()
		g.dijkstraRelax(curr)
	}
}

func (g *WeightedDirectedGraph[T]) acyclic(v T) {
	g.mark(false, true)
	g.distTo.Clear()
	g.pq = NewIndexHeap[T, float64](nil)
	g.pq.Update(v, 0)

	for node := range g.Sorted().Iterate() {
		g.dijkstraRelax(node)
	}
}

func (g *WeightedDirectedGraph[T]) bellmanFord(v T) {
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

func (g *WeightedDirectedGraph[T]) HasNegtiveCycle() bool {
	return g.hasNegtiveCycle
}

func (g *WeightedDirectedGraph[T]) ShortestPath(from, to T) typesw.IterableT[*Edge[T]] {
	g.Mark()
	if !g.Reachable(from, to) {
		return typesw.EmptyIterable[*Edge[T]]()
	}
	if g.negtiveEdge.Empty() {
		if g.hasCycle {
			g.dijkstra(from)
		} else {
			g.acyclic(from)
		}
	} else {
		g.bellmanFord(from)
	}
	// fmt.Println(g.edgeTo)
	// fmt.Println(g.distTo)
	s := NewStack(8)
	for curr := to; g.cmp(curr, from) != 0; {
		edge := g.edgeTo.Get(curr)
		s.Push(edge)
		// fmt.Println(edge, curr, g.edgeTo, edge.Other(curr, g.cmp))
		curr = edge.Other(curr, g.cmp)
	}
	// res = append(res, from)
	return typesw.ToIterable[*Edge[T]](s.Iter())
}
