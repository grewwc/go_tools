package containerW

type UF struct {
	id    []int
	sz    []int
	count int
}

func NewUF(n int) *UF {
	id := make([]int, n)
	for i := range id {
		id[i] = i
	}
	sz := make([]int, n)
	return &UF{id, sz, n}
}

func (uf *UF) Union(p, q int) {
	if !uf.Connected(p, q) {
		pid, qid := uf.root(p), uf.root(q)
		id, sz := uf.id, uf.sz
		if sz[pid] < sz[qid] {
			id[pid] = qid
			sz[qid] += sz[pid]
		} else {
			id[qid] = pid
			sz[pid] += sz[qid]
		}
		uf.count--
	}
}

func (uf *UF) Connected(p, q int) bool {
	return uf.root(p) == uf.root(q)
}

func (uf *UF) root(i int) int {
	id := uf.id 
	for id[i] != i {
		id[i] = id[id[i]]
		i = id[i]
	}
	return i
}

func (uf *UF) Components() int {
	return uf.count
}
