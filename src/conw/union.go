package conw

import "github.com/grewwc/go_tools/src/typew"

type UF[T comparable] struct {
	id          typew.IMap[T, T]
	sz          typew.IMap[T, int]
	group_count int
}

func NewUF[T comparable](nodes ...T) *UF[T] {
	id := NewMap[T, T]()
	sz := NewMap[T, int]()
	ret := &UF[T]{id, sz, 0}
	ret.AddNodes(nodes...)
	return ret
}

func (uf *UF[T]) Union(i, j T) {
	uf.AddNode(i)
	uf.AddNode(j)
	if !uf.IsConnected(i, j) {
		pid, qid := uf.root(i), uf.root(j)
		id, sz := uf.id, uf.sz
		sz1, sz2 := sz.Get(pid), sz.Get(qid)
		if sz1 < sz2 {
			id.Put(pid, qid)
			sz.Put(qid, sz1+sz2)
		} else {
			id.Put(qid, pid)
			sz.Put(pid, sz1+sz2)
		}
		uf.group_count--
	}
}

func (uf *UF[T]) AddNode(node T) bool {
	if uf.id.Contains(node) {
		return false
	}
	uf.id.Put(node, node)
	uf.group_count++
	return true
}

func (uf *UF[T]) AddNodes(nodes ...T) {
	for _, node := range nodes {
		uf.AddNode(node)
	}
}

func (uf *UF[T]) IsConnected(i, j T) bool {
	return uf.root(i) == uf.root(j)
}

func (uf *UF[T]) root(i T) T {
	id := uf.id
	for id.GetOrDefault(i, i) != i {
		id.Put(i, id.GetOrDefault(id.GetOrDefault(i, i), i))
		i = id.GetOrDefault(i, i)
	}
	return i
}

func (uf *UF[T]) NumNodes() int {
	return uf.id.Size()
}

func (uf *UF[T]) NumGroups() int {
	return uf.group_count
}

func (uf *UF[T]) Contains(node T) bool {
	return uf.id.Contains(node)
}
