package test

import (
	"math/rand"
	"testing"

	"github.com/grewwc/go_tools/src/cw"
)

var trie *cw.Trie

func generateRandomWord() []string {

	var letters []rune
	for i := 'a'; i <= 'z'; i++ {
		letters = append(letters, i)
	}

	res := make([]string, 0)
	for i := 0; i < 20; i++ {
		val := rand.Intn(26)
		str := ""
		if val == 0 {
			val++
		}
		for j := 0; j <= val; j++ {
			idx := rand.Intn(val)
			str += string(letters[idx])
		}
		res = append(res, str)
	}
	return res
}

func TestTrie(t *testing.T) {
	for i := 0; i < 1000; i++ {
		trie = cw.NewTrie()
		words := generateRandomWord()
		// add to trie
		for _, word := range words {
			trie.Insert(word)
		}

		// test
		for _, word := range words {
			if !trie.Contains(word) {
				t.Fatalf("%s not exist.", word)
			}
		}

		// delete one word
		trie.Delete(words[0])
		if trie.Contains(words[0]) {
			t.Logf("%v", words)
			t.Fatalf("%s should not exists.", words[0])
		}
	}
}
