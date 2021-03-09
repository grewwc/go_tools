package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

type bgTask struct {
	running  bool
	task     func()
	interval time.Duration
	done     chan struct{}
}

var backupFileTypes = containerW.NewSet(
	".go", ".py", ".cpp", ".tex", ".txt", ".htm",
	".bib", ".java", ".c", ".js", ".ts", ".html", ".css",
	".csv", ".xls", ".xlsx", ".out", ".jpg", ".jpeg", ".png",
)

var ignores = containerW.NewSet(
	".git", ".vscode", ".idea", "node_modules",
)

func newBGTask(task func(), interval time.Duration) *bgTask {
	return &bgTask{running: false,
		task:     task,
		done:     make(chan struct{}),
		interval: interval}
}

func (t bgTask) start() {
	if t.running {
		return
	}
	if t.task == nil {
		fmt.Println("task is nil !")
		return
	}
	t.running = true
	go func() {
		defer close(t.done)
		for {
			select {
			case <-t.done:
				return
			case <-time.Tick(t.interval):
				t.task()
			}
		}
	}()

	for {
		if _, ok := <-t.done; ok {
			time.Sleep(1)
		} else {
			return
		}
	}
}

func (t bgTask) stop() {
	if !t.running {
		return
	}
	t.done <- struct{}{}
}

func copyTo(from, to string) {
	if !utilsW.IsNewer(from, to) {
		return
	}
	fromDir := filepath.Dir(from)
	toDir := filepath.Dir(to)
	if !utilsW.IsExist(fromDir) {
		os.MkdirAll(fromDir, 0777)
	}
	if !utilsW.IsExist(toDir) {
		os.MkdirAll(toDir, 0777)
	}
	if err := utilsW.CopyFile(from, to); err != nil {
		log.Println(err)
	}
}

func task(fromRootDir, toRootDir string) {
	fromRootDir = filepath.ToSlash(fromRootDir)
	toRootDir = filepath.ToSlash(toRootDir)
	if !utilsW.IsDir(toRootDir) {
		log.Fatalf("%q is not a valid path\n", toRootDir)
	}
	q := containerW.NewQueue(fromRootDir)
	for !q.Empty() {
		cur := q.Dequeue().(string)
		for _, sub := range utilsW.LsDir(cur) {
			absPath := filepath.Join(cur, sub)
			absPath = filepath.ToSlash(absPath)
			ext := strings.ToLower(filepath.Ext(absPath))
			if utilsW.IsDir(absPath) {
				if !ignores.Contains(sub) {
					q.Enqueue(absPath)
				}
			} else if backupFileTypes.Contains(ext) {
				// fmt.Println("here", absPath, fromRootDir)
				relPath := stringsW.StripPrefix(absPath, fromRootDir)
				// fmt.Println("from", absPath, "to", filepath.Join(toRootDir, relPath))
				copyTo(absPath, filepath.Join(toRootDir, relPath))
			}
		}
		// fmt.Println(cur)
	}
}

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	interval := time.Minute

	fs.Int("s", 30, "interval in seconds")
	fs.Int("m", 10, "interval in minutes")
	fs.Int("H", 1, "interval in hours")
	fs.String("from", "./", "source folder")
	fs.String("to", "", "dest folder")

	parsedResults := terminalW.ParseArgsCmd()
	if parsedResults == nil || parsedResults.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}

	from, err := parsedResults.GetFlagVal("from")
	if err != nil {
		from = "./"
	}
	to, err := parsedResults.GetFlagVal("to")
	if err != nil {
		log.Fatalln(err)
	}

	if n := parsedResults.GetNumArgs(); n != -1 {
		interval = time.Second * time.Duration(n)
		goto skipTo
	}
	if parsedResults.ContainsFlagStrict("H") {
		val, _ := parsedResults.GetFlagVal("H")
		n, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalln(err)
		}
		interval = time.Hour * time.Duration(n)
	}

	if parsedResults.ContainsFlagStrict("m") {
		val, _ := parsedResults.GetFlagVal("m")
		n, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalln(err)
		}
		interval = time.Minute * time.Duration(n)
	}

	if parsedResults.ContainsFlagStrict("s") {
		val, _ := parsedResults.GetFlagVal("s")
		n, err := strconv.Atoi(val)
		if err != nil {
			log.Fatalln(err)
		}
		interval = time.Second * time.Duration(n)
	}

skipTo:
	t := newBGTask(func() {
		task(from, to)
	}, interval)
	t.start()
}