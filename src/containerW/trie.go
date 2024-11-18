package containerW

type Trie struct {
	root map[rune]*Trie
	end  map[rune]int
}

/** Initialize your data structure here. */
func NewTrie() *Trie {
	return &Trie{end: make(map[rune]int), root: make(map[rune]*Trie)}
}

/** Inserts a word into the trie. */
func (t *Trie) Insert(word string) {
	cur := t
	for cnt, ch := range word {
		if _, exists := cur.root[ch]; !exists {
			newTrie := NewTrie()
			cur.root[ch] = newTrie
		}
		if cnt+1 == len(word) {
			cur.end[ch] += 1
		}
		cur = cur.root[ch]
	}
}

/** Returns if the word is in the trie. */
func (t *Trie) Contains(word string) bool {
	if len(word) == 0 {
		return false
	}
	cur := t
	for cnt, ch := range word {
		if _, exists := cur.root[ch]; !exists {
			return false
		}
		if cnt+1 == len(word) {
			return t.end[ch] > 0
		}
		cur = cur.root[ch]
	}
	return true
}

/** Returns if the word is in the trie. */
func (t *Trie) LooseContains(word string) bool {
	if len(word) == 0 {
		return false
	}
	cur := t
	for _, ch := range word {
		if _, exists := cur.root[ch]; !exists {
			return false
		}
		cur = cur.root[ch]
	}
	return true
}

func (t *Trie) Delete(word string) bool {
	if !t.Contains(word) {
		return false
	}
	cur := t
	for _, ch := range word {
		cur.end[ch]--
		if cur.end[ch] == 0 {
			delete(cur.root, ch)
		}
	}
	return true
}
