package utilsW

import (
	"log"
	"os"
	"path/filepath"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/strW"
)

var trie *containerW.Trie

const (
	filename = "google-10000-english.txt"
)

func init() {
	wordFile := filepath.Join(GetDirOfTheFile(), filename)
	trie = containerW.NewTrie()
	f, err := os.Open(wordFile)
	if err != nil {
		log.Fatalln(err)
	}
	for line := range strW.SplitByToken(f, "\n", false) {
		trie.Insert(line)
	}
}

func IsWord(word string) bool {
	return trie.Contains(word)
}
