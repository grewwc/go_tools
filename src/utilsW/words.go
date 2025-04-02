package utilsW

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/strW"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var trie *containerW.Trie

const (
	filename = "words.txt"
)

func init() {
	wordFile := filepath.Join(GetDirOfTheFile(), filename)
	trie = containerW.NewTrie()
	f, err := os.Open(wordFile)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strW.IsBlank(line) {
			continue
		}
		trie.Insert(line)
		lower := strings.ToLower(line)
		if lower != line {
			trie.Insert(lower)
		}
		title := cases.Title(language.English, cases.Compact).String(line)
		if title != line {
			trie.Insert(title)
		}
		upper := strings.ToUpper(line)
		if line != upper {
			trie.Insert(upper)
		}
	}
}

func IsWord(word string) bool {
	return trie.Contains(word)
}

func ShowPrefix(word string, n int) []string {
	return trie.ShowPrefix(word, n)
}
