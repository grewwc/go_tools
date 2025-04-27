package cw

import "github.com/grewwc/go_tools/src/typesw"

type UF[T any] struct {
	id   typesw.IMap[T, T]
	root typesw.IMap[T, *TreeSet[T]]
	cmp  typesw.CompareFunc[T]
}

func NewUF[T any](cmp typesw.CompareFunc[T], nodes ...T) *UF[T] {
	id := NewMap[T, T]()
	root := NewMap[T, *TreeSet[T]]()
	if cmp == nil {
		cmp = typesw.CreateDefaultCmp[T]()
	}
	ret := &UF[T]{id, root, cmp}
	ret.AddNodes(nodes...)
	return ret
}

func (uf *UF[T]) Union(i, j T) bool {
	uf.AddNode(i)
	uf.AddNode(j)
	if uf.IsConnected(i, j) {
		return false
	}
	pid, qid := uf.find(i), uf.find(j)
	sz1, sz2 := uf.root.Get(pid).Size(), uf.root.Get(qid).Size()
	if sz1 < sz2 {
		uf.id.Put(pid, qid)
		uf.root.Get(qid).Union(uf.root.Get(pid))
		uf.root.Delete(pid)
	} else {
		uf.id.Put(qid, pid)
		uf.root.Get(pid).Union(uf.root.Get(qid))
		uf.root.Delete(qid)
	}
	return true
}

func (uf *UF[T]) AddNode(node T) bool {
	if uf.id.Contains(node) {
		return false
	}
	uf.id.Put(node, node)
	set := NewTreeSet(uf.cmp)
	set.Add(node)
	uf.root.Put(node, set)
	return true
}

func (uf *UF[T]) AddNodes(nodes ...T) {
	for _, node := range nodes {
		uf.AddNode(node)
	}
}

func (uf *UF[T]) IsConnected(i, j T) bool {
	return uf.cmp(uf.find(i), uf.find(j)) == 0
}

func (uf *UF[T]) find(i T) T {
	id := uf.id
	root := i
	for uf.cmp(id.GetOrDefault(root, root), root) != 0 {
		root = id.GetOrDefault(root, root)
	}
	// path compression
	for uf.cmp(i, root) != 0 {
		parent := id.GetOrDefault(i, i)
		id.Put(i, root)
		i = parent
	}
	return i
}

func (uf *UF[T]) NumNodes() int {
	return uf.id.Size()
}

func (uf *UF[T]) NumGroups() int {
	return uf.root.Size()
}

func (uf *UF[T]) Contains(node T) bool {
	return uf.id.Contains(node)
}
