package terminalW

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"golang.org/x/tools/godoc/util"
)

var DefaultExtensions = [...]string{".py", ".cpp", ".js", ".txt", ".h", ".c", ".tex", ".html", ".css", ".java", ".go", ".cc"}
var Extensions string
var CheckExtension bool
var CheckFileWithoutExt bool

var NumPrint int64 = 3
var Count int64

// maximum 5000 threads
var maxThreads = make(chan struct{}, 5000)
var Verbose bool
var CountMu sync.Mutex

// how many levels to search
var MaxLevel int32

func isTextFile(filename string) bool {
	b, err := ioutil.ReadFile(filename)
	if err != nil && Verbose {
		fmt.Fprintln(os.Stderr, err)
	}
	return util.IsText(b)
}

// this function is the main part
// acts like a framework
func Find(rootDir string, task func(string), wg *sync.WaitGroup, level int32) {
	defer wg.Done()
	if atomic.LoadInt32(&level) > MaxLevel {
		return
	}
	maxThreads <- struct{}{}
	defer func() { <-maxThreads }()
	CountMu.Lock()
	if Count >= NumPrint {
		CountMu.Unlock()
		return
	}
	CountMu.Unlock()
	subs, err := ioutil.ReadDir(rootDir)
	if err != nil {
		if Verbose {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	for _, sub := range subs {
		subName := path.Join(rootDir, sub.Name())
		extName := path.Ext(subName)
		if sub.IsDir() {
			wg.Add(1)
			go Find(subName, task, wg, atomic.AddInt32(&level, 1))
			atomic.AddInt32(&level, -1)
		} else if !CheckExtension {
			// read the file content to check if is human readable
			if extName == "" && !isTextFile(subName) {
				continue
			}
			task(subName)
		} else if strings.Contains(Extensions, extName) {
			if extName == "" && (!CheckFileWithoutExt || !isTextFile(subName)) {
				continue
			}
			task(subName)
		}
	}
}
