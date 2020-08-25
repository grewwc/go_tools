package terminalW

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"
)

var DefaultExtensions = [...]string{".py", ".cpp", ".js", ".txt", ".h", ".c", ".tex", ".html", ".css", ".java", ".go", ".cc"}
var Extensions string
var CheckExtension bool = true

var NumPrint int64 = 10
var Count int64

// maximum 5000 threads
var maxThreads = make(chan struct{}, 5000)
var Verbose bool
var CountMu sync.Mutex

// how many levels to search
var MaxLevel int32

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
		} else if (!CheckExtension) || (extName != "" && strings.Contains(Extensions, extName)) {
			task(subName)
		}
	}
}
