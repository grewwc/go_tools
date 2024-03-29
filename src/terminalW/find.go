package terminalW

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/utilsW"
)

var Once sync.Once

var FileNamesToCheck = containerW.NewSet()
var FileNamesNOTCheck = containerW.NewSet()

var Extensions = containerW.NewSet()
var CheckExtension bool
var Exclude bool

var NumPrint int64 = 5

var Count int64

// maximum 4 threads
var maxThreads = make(chan struct{}, 4)
var Verbose bool
var CountMu sync.Mutex

// how many levels to search
var MaxLevel int32

func ChangeThreads(num int) {
	close(maxThreads)
	maxThreads = make(chan struct{}, num)
	log.Println("change threads num to", num)
}

// this function is the main part
// acts like a framework
func Find(rootDir string, task func(string), wg *sync.WaitGroup, level int32) {
	defer wg.Done()
	if level > MaxLevel {
		return
	}
	maxThreads <- struct{}{}
	defer func() { <-maxThreads }()
	CountMu.Lock()
	if Count >= NumPrint {
		CountMu.Unlock()
		Once.Do(func() {
			summaryString := utilsW.Sprintf("%d matches found\n", Count)
			utilsW.Println(strings.Repeat("-", len(summaryString)))
			matches := int64(math.Min(float64(Count), float64(NumPrint)))
			utilsW.Printf("%v matches found\n", matches)
		})
		os.Exit(0)
		return
	}
	CountMu.Unlock()
	subs, err := ioutil.ReadDir(rootDir)
	if err != nil {
		if Verbose {
			utilsW.Fprintln(os.Stderr, err)
		}
		return
	}

	for _, sub := range subs {
		subName := path.Join(rootDir, sub.Name())
		extName := path.Ext(subName)
		if (!FileNamesToCheck.Empty() && !FileNamesToCheck.Contains(filepath.Base(subName))) ||
			FileNamesNOTCheck.Contains(filepath.Base(subName)) {
			continue
		}
		if sub.IsDir() {
			wg.Add(1)
			go Find(subName, task, wg, level+1)
			continue
		}

		if !utilsW.IsTextFile(subName) {
			continue
		}

		if !CheckExtension {
			task(subName)
		} else if Extensions.Contains(extName) {
			task(subName)
		}
	}
}
