package cw

import (
	"github.com/grewwc/go_tools/src/typesw"
)

type zSetEntry[Key any] struct {
	Key   Key
	Score float64
}

type ZSet[Key any] struct {
	tree *RbTree[*zSetEntry[Key]]

	map_ *Map[Key, *zSetEntry[Key]]
}

func NewZSet[Key any]() *ZSet[Key] {
	cmp := func(entry1, entry2 *zSetEntry[Key]) int {
		diff := entry1.Score - entry2.Score
		if diff > 1e-5 {
			return 1
		}
		if diff < -1e-5 {
			return -1
		}
		return 0
	}
	return &ZSet[Key]{
		tree: NewRbTree(cmp),

		map_: NewMap[Key, *zSetEntry[Key]](),
	}
}

func (z *ZSet[Key]) Add(key Key, score float64) bool {
	if z.map_.Contains(key) {
		return false
	}
	newEntry := &zSetEntry[Key]{Key: key, Score: score}
	z.tree.Insert(newEntry)
	z.map_.Put(key, newEntry)
	return true
}

func (z *ZSet[Key]) UpdateScore(key Key, score float64) bool {
	if !z.map_.Contains(key) {
		return false
	}
	entry := z.map_.Get(key)
	entry.Score = score

	// re-insert
	z.tree.Delete(entry)
	z.tree.Insert(entry)
	return true
}

func (z *ZSet[Key]) Delete(key Key) bool {
	if !z.map_.Contains(key) {
		return false
	}
	entry := z.map_.Get(key)
	z.Delete(key)
	z.tree.Delete(entry)
	return true
}

func (z *ZSet[Key]) Len() int {
	return z.map_.Size()
}

func (z *ZSet[Key]) SearchRange(scoreLow, scoreHi float64) typesw.IterableT[*zSetEntry[Key]] {
	lo := &zSetEntry[Key]{
		Score: scoreLow,
	}
	hi := &zSetEntry[Key]{
		Score: scoreHi,
	}
	return z.tree.SearchRange(lo, hi)
}

// Rank returns -1 if not exists
func (z *ZSet[Key]) Rank(key Key) int {
	if !z.map_.Contains(key) {
		return -1
	}
	entry := z.map_.Get(key)
	min := z.tree.min
	cnt := 0
	for range z.SearchRange((*min).Score, entry.Score).Iterate() {
		cnt++
	}
	return cnt
}

func (z *ZSet[Key]) Min() *zSetEntry[Key] {
	if z.map_.Size() == 0 {
		return nil
	}
	return *z.tree.min
}

func (z *ZSet[Key]) Max() *zSetEntry[Key] {
	if z.map_.Size() == 0 {
		return nil
	}
	return *z.tree.max
}

func (z *ZSet[Key]) Score(key Key) float64 {
	if !z.map_.Contains(key) {
		return -1
	}
	return z.map_.Get(key).Score
}

func (z *ZSet[Key]) Contains(key Key) bool {
	return z.map_.Contains(key)
}

func (z *ZSet[Key]) Iter() typesw.IterableT[*zSetEntry[Key]] {
	return z.tree.Iter()
}

func (z *ZSet[Key]) Intersect(another *ZSet[Key]) *ZSet[Key] {
	if another == nil {
		return nil
	}
	result := NewZSet[Key]()
	for k := range z.Iter().Iterate() {
		if another.Contains(k.Key) {
			result.Add(k.Key, k.Score)
		}
	}
	return result
}

func (z *ZSet[Key]) Union(another *ZSet[Key]) *ZSet[Key] {
	result := NewZSet[Key]()
	for k := range z.Iter().Iterate() {
		result.Add(k.Key, k.Score)
	}
	if another == nil {
		return result
	}
	for k := range another.Iter().Iterate() {
		result.Add(k.Key, k.Score)
	}
	return result
}

func (z *ZSet[Key]) Subtract(another *ZSet[Key]) {
	if another == nil {
		return
	}
	for k := range another.Iter().Iterate() {
		z.Delete(k.Key)
	}
}

func (z *ZSet[Key]) Clear() {
	z.map_.Clear()
	z.tree.Clear()
}
