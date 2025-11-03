package cw

import (
	"fmt"
	"math"
	"unicode/utf8"
)

type Trie struct {
	children map[rune]*Trie
	// end  map[rune]bool
	cnt   int
	isEnd bool

	count int
}

/** Initialize your data structure here. */
func NewTrie() *Trie {
	return &Trie{children: make(map[rune]*Trie)}
}

/** Inserts a word into the trie. */
func (t *Trie) Insert(word string) error {
	cur := t
	inserted := false
	for cnt, ch := range word {
		var child *Trie
		var exists bool
		if child, exists = cur.children[ch]; !exists {
			child = NewTrie()
			inserted = true
			cur.children[ch] = child
		}
		chLen := utf8.RuneLen(ch)
		if chLen == -1 {
			return fmt.Errorf("failed to insert word: %s (invalid utf8 rune)", word)
		}
		child.cnt++
		if cnt+chLen == len(word) {
			child.isEnd = true
		}
		cur = cur.children[ch]
	}
	if inserted {
		t.count++
	}
	return nil
}

/** Returns if the word is in the trie. */
func (t *Trie) Contains(word string) bool {
	if len(word) == 0 {
		return true
	}
	cur := t
	for cnt, ch := range word {
		var child *Trie
		var exists bool
		if child, exists = cur.children[ch]; !exists {
			return false
		}
		chLen := utf8.RuneLen(ch)
		if chLen == -1 {
			return false
		}
		if cnt+chLen == len(word) {
			return child.isEnd
		}
		cur = child
	}
	return true
}

/** Returns if the word is in the trie. */
func (t *Trie) HasPrefix(word string) bool {
	if len(word) == 0 {
		return true
	}
	cur := t
	for _, ch := range word {
		var child *Trie
		var exists bool
		if child, exists = cur.children[ch]; !exists {
			return false
		}
		cur = child
	}
	return true
}

func (t *Trie) Delete(word string) bool {
	cur := t
	for cnt, ch := range word {
		var child *Trie
		var exists bool
		if child, exists = cur.children[ch]; !exists {
			return false
		}
		chLen := utf8.RuneLen(ch)
		if chLen == -1 {
			return false
		}
		if cnt+chLen == len(word) {
			if child.cnt <= 1 {
				delete(cur.children, ch)
			}
			child.isEnd = false
			break
		}
		cur = child
	}
	t.count--
	return true
}

func showPrefixHelper(t *Trie, prefix string, n int, isEnd bool) []string {
	if len(prefix) == 0 || n == 0 {
		return nil
	}
	res := make([]string, 0, n)
	if n == 0 {
		return nil
	}
	s := NewQueue[*Tuple]()
	curr := prefix
	currTrie := t
	s.Enqueue(NewTuple(currTrie, curr, isEnd))
	for !s.Empty() {
		currTuple := s.Dequeue()
		currTrie, curr, isEnd = currTuple.Get(0).(*Trie), currTuple.Get(1).(string), currTuple.Get(2).(bool)
		if curr != "" && isEnd {
			res = append(res, curr)
			n--
			if n <= 0 {
				goto end
			}
		}
		for ch, subT := range currTrie.children {
			if n > 0 {
				s.Enqueue(NewTuple(subT, curr+string(ch), subT.isEnd))
			} else {
				goto end
			}
		}
	}
end:
	return res
}

func (t *Trie) ShowPrefix(prefix string, totalNum int) []string {
	if len(prefix) == 0 || totalNum == 0 {
		return nil
	}
	if totalNum < 0 {
		totalNum = math.MaxInt
	}
	var ch rune
	var exists bool
	var prefixInDict bool
	// find next prefix
	for _, ch = range prefix {
		if t, exists = t.children[ch]; !exists {
			return nil
		}
		prefixInDict = t.isEnd
	}
	return showPrefixHelper(t, prefix, totalNum, prefixInDict)
}

func (t *Trie) Len() int {
	return t.count
}

func (t *Trie) Empty() bool {
	return t.count <= 0
}
