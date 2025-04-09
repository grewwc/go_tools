package conw

import "math"

type Trie struct {
	data map[rune]*Trie
	end  map[rune]bool
}

/** Initialize your data structure here. */
func NewTrie() *Trie {
	return &Trie{end: make(map[rune]bool), data: make(map[rune]*Trie)}
}

/** Inserts a word into the trie. */
func (t *Trie) Insert(word string) {
	cur := t
	for cnt, ch := range word {
		if _, exists := cur.data[ch]; !exists {
			newTrie := NewTrie()
			cur.data[ch] = newTrie
		}
		if cnt+len(string(ch)) == len(word) {
			cur.end[ch] = true
		}
		cur = cur.data[ch]
	}
}

/** Returns if the word is in the trie. */
func (t *Trie) Contains(word string) bool {
	if len(word) == 0 {
		return false
	}
	cur := t
	for cnt, ch := range word {
		if _, exists := cur.data[ch]; !exists {
			return false
		}
		if cnt+len(string(ch)) == len(word) {
			return cur.end[ch]
		}
		cur = cur.data[ch]
	}
	return true
}

/** Returns if the word is in the trie. */
func (t *Trie) HasPrefix(word string) bool {
	if len(word) == 0 {
		return false
	}
	cur := t
	for _, ch := range word {
		if _, exists := cur.data[ch]; !exists {
			return false
		}
		cur = cur.data[ch]
	}
	return true
}

func (t *Trie) Delete(word string) bool {
	// if !t.Contains(word) {
	// 	return false
	// }
	cur := t
	for cnt, ch := range word {
		if _, ok := cur.data[ch]; !ok {
			return false
		}
		if cnt+len(string(ch)) == len(word) {
			delete(cur.end, ch)
			delete(cur.data, ch)
		}
		cur = cur.data[ch]
	}
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
	s := NewQueue()
	curr := prefix
	currTrie := t
	s.Enqueue(NewTuple(currTrie, curr, isEnd))
	for !s.Empty() {
		currTuple := s.Dequeue().(*Tuple)
		currTrie, curr, isEnd = currTuple.Get(0).(*Trie), currTuple.Get(1).(string), currTuple.Get(2).(bool)
		if curr != "" && isEnd {
			res = append(res, curr)
			n--
			if n <= 0 {
				goto end
			}
		}
		for ch, subT := range currTrie.data {
			if n > 0 {
				s.Enqueue(NewTuple(subT, curr+string(ch), currTrie.end[ch]))
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
		if t, exists = t.data[ch]; !exists {
			return nil
		}
		prefixInDict = t.end[ch]
	}
	return showPrefixHelper(t, prefix, totalNum, prefixInDict)
}
