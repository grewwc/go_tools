package containerW

type Trie struct {
	end    map[byte]bool
	values map[byte]*Trie
}

/** Initialize your data structure here. */
func NewTrie() *Trie {
	return &Trie{end: make(map[byte]bool), values: make(map[byte]*Trie)}
}

/** Inserts a word into the trie. */
func (this *Trie) Insert(word string) {
	this.insertFrom(&word, 0)
}

/** Returns if the word is in the trie. */
func (this *Trie) Search(word string) bool {
	return this.searchFrom(&word, 0, true)
}

/** Returns if the word is in the trie. */
func (this *Trie) LooseSearch(word string) bool {
	return this.searchFrom(&word, 0, false)
}

/** Returns if there is any word in the trie that starts with the given prefix. */
func (this *Trie) StartsWith(prefix string) bool {
	return this.startsWithFrom(&prefix, 0)
}

func (this *Trie) Delete(word string) bool {
	var record []byte
	var path []*Trie
	return this.deleteFrom(&word, &record, &path, 0)
}

// private mumber functions
func (this *Trie) startsWithFrom(prefix *string, from int) bool {
	if len(*prefix) == from {
		return true
	}

	if subTrie, exist := this.values[(*prefix)[from]]; exist {
		return subTrie.startsWithFrom(prefix, from+1)
	}

	return false
}

func (this *Trie) insertFrom(word *string, from int) {
	if len(*word) == from {
		return
	}
	curByte := (*word)[from]
	if len(*word) == from+1 {
		this.end[curByte] = true
		if _, exist := this.values[curByte]; !exist {
			newTrie := NewTrie()
			this.values[curByte] = newTrie
		}
		return
	}

	if subTrie, exist := this.values[curByte]; exist {
		subTrie.insertFrom(word, from+1)
	} else {
		newTrie := NewTrie()
		this.values[curByte] = newTrie
		newTrie.insertFrom(word, from+1)
	}
}

func (this *Trie) searchFrom(word *string, from int, strict bool) bool {
	if len(*word) == from {
		return true
	}
	curByte := (*word)[from]
	if len(*word) == from+1 {
		if _, exist := this.values[curByte]; exist {
			if strict {
				return this.end[curByte]
			}
			return true
		}
		return false
	}

	if subTrie, exist := this.values[curByte]; exist {
		return subTrie.searchFrom(word, from+1, strict)
	}

	return false
}

func (this *Trie) deleteFrom(word *string, record *[]byte, path *[]*Trie, from int) bool {
	if len(*word) == from {
		return true
	}

	curByte := (*word)[from]
	if len(*word) == from+1 {
		if subTrie, exist := this.values[curByte]; exist {
			if this.end[curByte] {
				if subTrie == nil {
					delete(this.values, curByte)
				}
				this.end[curByte] = false
				// delete according to "record" and "path"
				for i := len(*path) - 1; i >= 0; i-- {
					curByte = (*record)[i]
					curTrie := (*path)[i]
					if len(curTrie.values) == 1 { // only has "curByte" subtrie
						if curTrie.end[curByte] {
							return true
						}
						delete(curTrie.values, curByte)
					}
				}
				return true
			}
			return false
		}
		return false
	}

	if subTrie, exist := this.values[curByte]; exist {
		// fmt.Println("here", curByte)
		*record = append(*record, curByte)
		*path = append(*path, this)
		return subTrie.deleteFrom(word, record, path, from+1)
	}
	return false
}
