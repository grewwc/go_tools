package containerW

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
			cur.end[ch] = false
			delete(cur.data, ch)
		}
		cur = cur.data[ch]
	}
	return true
}

func showPrefixHelper(t *Trie, prefix string, n int) []string {
	if len(prefix) == 0 || n == 0 {
		return nil
	}
	res := make([]string, 0, n)
	if n == 0 {
		return nil
	}
	s := NewQueue()
	curr := prefix
	currMap := t.data
	s.Enqueue(NewTuple(currMap, curr))
	for !s.Empty() {
		currTuple := s.Dequeue().(*Tuple)
		currMap, curr = currTuple.Get(0).(map[rune]*Trie), currTuple.Get(1).(string)
		if curr != "" {
			res = append(res, curr)
			n--
		}
		for ch, t := range currMap {
			if n > 0 {
				s.Enqueue(NewTuple(t.data, curr+string(ch)))
			} else {
				break
			}
		}
	}
	return res
}

func (t *Trie) ShowPrefix(prefix string, n int) []string {
	if len(prefix) == 0 || n == 0 {
		return nil
	}
	if n < 0 {
		n = 128
	}
	// find next prefix
	for _, ch := range prefix {
		if _, exists := t.data[ch]; !exists {
			return nil
		}
		t = t.data[ch]
	}
	return showPrefixHelper(t, prefix, n)
}
