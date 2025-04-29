package terminalw

import (
	"fmt"
	"log"
	"math"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/grewwc/go_tools/src/cw"
	utilsw "github.com/grewwc/go_tools/src/utilsw"
)

var Once sync.Once

var FileNamesToCheck = cw.NewSet()
var FileNamesNOTCheck = cw.NewSet()

var Extensions = cw.NewSet()
var CheckExtension bool
var Exclude bool

var NumPrint int64 = 5

var Count int64

// maximum 4 threads
var maxThreads = make(chan struct{}, 4)
var Verbose bool

// how many levels to search
var MaxLevel int32

func ChangeThreads(num int) {
	close(maxThreads)
	maxThreads = make(chan struct{}, num)
	log.Println("change threads num to", num)
}

func Find(rootDir string, task func(string), wg *sync.WaitGroup, level int32) {
	defer wg.Done()
	if level > MaxLevel {
		return
	}
	maxThreads <- struct{}{}
	defer func() { <-maxThreads }()
	if atomic.LoadInt64(&Count) >= NumPrint {
		Once.Do(func() {
			summaryString := fmt.Sprintf("%d matches found\n", Count)
			fmt.Println(strings.Repeat("-", len(summaryString)))
			matches := int64(math.Min(float64(Count), float64(NumPrint)))
			fmt.Printf("%v matches found\n", matches)
		})
		os.Exit(0)
		return
	}
	subs, err := os.ReadDir(rootDir)
	if err != nil {
		if Verbose {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}

	for _, sub := range subs {
		subName := path.Join(rootDir, sub.Name())
		extName := path.Ext(subName)
		baseName := path.Base(subName)
		if (!FileNamesToCheck.Empty() && !FileNamesToCheck.Contains(baseName)) || FileNamesNOTCheck.Contains(baseName) {
			continue
		}
		if sub.IsDir() {
			wg.Add(1)
			go Find(subName, task, wg, level+1)
			continue
		}

		if !utilsw.IsTextFile(subName) {
			continue
		}

		if !CheckExtension {
			task(subName)
		} else if Extensions.Contains(extName) {
			task(subName)
		}
	}
}
