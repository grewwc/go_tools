package syncDir

import (
	"fmt"
	"go_tools/src/configW"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	root = ".syncmaps"
)

var syncdir string

var variablesMap map[string]string
var dirMap map[string]string
var parsedAttrsMap map[string][]string

var count int64

// func initSupportedAttrs() {
// 	supportedAttrs = make(map[string]bool)
// 	supportedAttrs["ignore"] = true
// }

func init() {
	homedir := os.Getenv("HOME")
	if homedir == "" {
		log.Fatal("home dir is empty")
	}

	syncdir = filepath.Join(homedir, root)

	parsedResults := configW.Parse(syncdir)
	variablesMap = parsedResults.Variables
	dirMap = parsedResults.Mapping
	parsedAttrsMap = parsedResults.Attributes

	// replace glob for dirmap
	replaceGlobForDirmap()

	// replace glob for parsedattrs
	replaceGlobForIgnore()
}

func replaceGlobForDirmap() {
	for src, dest := range dirMap {
		matches, err := filepath.Glob(src)
		if err != nil {
			log.Println(err)
		}

		// src is not in glob form
		if len(matches) == 1 && matches[0] == src {
			continue
		}

		delete(dirMap, src)
		for _, match := range matches {
			dirMap[match] = dest
		}
	}
}

func replaceGlobForIgnore() {
	var replaced []string
	for attr, dirs := range parsedAttrsMap {
		for _, dir := range dirs {
			matches, err := filepath.Glob(dir)
			if err != nil {
				log.Println(err)
			}
			for _, match := range matches {
				replaced = append(replaced, filepath.ToSlash(match))
			}
		}
		parsedAttrsMap[attr] = replaced
	}
}

func copyFile(from, to string, wg *sync.WaitGroup) {
	defer wg.Done()
	var overwight bool
	ffrom, err := os.Open(from)
	if err != nil {
		// log.Println(err)
		return
	}
	defer ffrom.Close()

	fromUnixNano := getFileUnixNano(from)
	if fromUnixNano == -1 {
		log.Printf("copy %q error\n", clean(from))
		return
	}

	if isDir(to) {
		to = filepath.Join(to, filepath.Base(from))
	}

	var toUnixNano int64
	if !isExist(to) {
		toUnixNano = fromUnixNano - 1
	} else {
		toUnixNano = getFileUnixNano(to)
		overwight = true
	}

	// if "to" is newer, ignore
	if fromUnixNano < toUnixNano {
		// fmt.Println("early", from, to)
		return
	}

	// fmt.Println(fromUnixNano, toUnixNano)

	fto, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY, 0744)
	if err != nil {
		// fmt.Println("what")
		log.Println(err)
	}
	defer fto.Close()

	_, err = io.Copy(fto, ffrom)
	if err != nil {
		log.Println(err)
	}

	atomic.AddInt64(&count, 1)
	if overwight {
		fmt.Printf("[Overwright!!] copy %q to %q\n", clean(from), clean(to))
	} else {
		fmt.Printf("copy %q to %q\n", clean(from), clean(to))
	}
}

func copyDir(src, dest string, wg *sync.WaitGroup) {
	defer wg.Done()
	if shouldIgnoreDir(src) {
		// fmt.Println("here", src)
		return
	}

	if isDir(src) {
		subs, err := ioutil.ReadDir(src)
		if err != nil {
			log.Println(err)
			return
		}

		// if dest directory is not exist, create one
		_, err = os.Stat(dest)
		if os.IsNotExist(err) {
			err := os.MkdirAll(dest, 0644)
			if err != nil {
				log.Println(err)
			}
		}
		for _, sub := range subs {
			wg.Add(1)
			subDestDir := filepath.Join(dest, sub.Name())
			subSrcDir := filepath.Join(src, sub.Name())
			// fmt.Println("here", subDestDir)
			go copyDir(subSrcDir, subDestDir, wg)
		}
	} else if isRegular(src) {
		wg.Add(1)
		go copyFile(src, dest, wg)
	}
}

func run() {

	var wg sync.WaitGroup
	for from, to := range dirMap {
		wg.Add(1)
		go copyDir(from, to, &wg)
	}

	// fmt.Println(dirMap, "\n\n")
	// fmt.Println(variablesMap, "\n\n")
	// fmt.Println(parsedAttrsMap, "\n\n")
	wg.Wait()
}

func Main() {
	fmt.Printf("  put config files in: %q\n", clean(syncdir))
	info := "  attribute files are in format *.attr "
	fmt.Println(info)
	fmt.Println(" ", strings.Repeat("-", len(info)))
	run()
	fmt.Printf("  %v files copied\n", count)
}
