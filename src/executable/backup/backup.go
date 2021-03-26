package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	backupTypeFile = ".backup-type"
)

var remote bool

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
	".bib",
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
	t.task()
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

// isNewer return True if local is newer than remote
func isNewer(local, remote string) bool {
	cmd := exec.Command("ssh", "wwc129@147.8.146.85", "stat", "-c", "%Y", remote)

	// fmt.Println(cmd.Args)
	var out bytes.Buffer
	var e bytes.Buffer

	cmd.Stdin = os.Stdin
	cmd.Stdout = &out
	cmd.Stderr = &e
	if err := cmd.Run(); err != nil {
		return true
	}
	if e.Len() != 0 {
		return true
	}
	remoteDateSec, err := strconv.Atoi(strings.TrimSpace(out.String()))
	if err != nil {
		return true
	}
	// fmt.Println(remoteDateSec)

	info, err := os.Stat(local)
	if err != nil {
		log.Fatalln(err)
	}
	localModTime := info.ModTime().UnixNano() / 1e9
	return int(localModTime) > remoteDateSec
}

// copyTo should handle remote scp
func copyTo(from, to string) {
	if !remote && !utilsW.IsNewer(from, to) {
		return
	}
	to = filepath.ToSlash(to)
	fromDir := filepath.Dir(from)
	toDir := filepath.Dir(to)
	if !utilsW.IsExist(fromDir) {
		os.MkdirAll(fromDir, 0777)
	}
	if !remote && !utilsW.IsExist(toDir) {
		os.MkdirAll(toDir, 0777)
	}
	if !remote {
		if err := utilsW.CopyFile(from, to); err != nil {
			log.Println(err)
		}
	} else {
		if isNewer(from, to) {
			// cmd := exec.Command("scp", "main.go", "wwc129@147.8.146.85:~/")
			// out, err := cmd.CombinedOutput()
			// if err != nil {
			// 	log.Println(err)
			// }
			// print(string(out))
			fmt.Println(from)
			fmt.Println(to)
		}
	}
	now := time.Now().Format("2006/01/02 15:04:05")
	fmt.Printf("  %s: %s => %s\n", color.HiWhiteString(now),
		color.GreenString(from), color.GreenString(to))
}

func task(fromRootDir, toRootDir string) {
	fromRootDir = filepath.ToSlash(fromRootDir)
	toRootDir = filepath.ToSlash(toRootDir)
	if !remote && !utilsW.IsDir(toRootDir) {
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

func init() {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		log.Println("cannot get $HOME")
		return
	}
	fname := filepath.Join(homeDir, backupTypeFile)

	func() {
		b, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Println(err)
		}
		content := string(b)
		for _, line := range strings.Split(content, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if line[0] != '.' {
				line = "." + line
			}
			backupFileTypes.Add(line)
		}
	}()
}

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	interval := time.Minute

	fs.Int("s", 30, "interval in seconds")
	fs.Int("m", 10, "interval in minutes")
	fs.Int("H", 1, "interval in hours")
	fs.String("from", "./", "source folder")
	fs.String("to", "", "dest folder")
	fs.Bool("watch", false, "keep watching folder changes")
	fs.Bool("remote", false, "scp to remote")

	parsedResults := terminalW.ParseArgsCmd("watch", "remote")
	if parsedResults == nil || parsedResults.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		fmt.Printf("You can define more types in %q\n", "$HOME/.backup-type")
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

	fmt.Printf("copy from: %s to: %s\n", color.YellowString(from), color.YellowString(to))

	remote = parsedResults.GetBooleanArgs().Contains("remote")

	if !parsedResults.GetBooleanArgs().Contains("watch") {
		task(from, to)
		return
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
