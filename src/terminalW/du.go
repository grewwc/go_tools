package terminalW

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

var threadControl = make(chan struct{}, 50)

func listFile(path string) []os.FileInfo {
	threadControl <- struct{}{}
	defer func() { <-threadControl }()
	fileInfos, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil
	}
	return fileInfos
}

func walkDir(root string, fileSize chan<- int64, wg *sync.WaitGroup) {
	defer wg.Done()
	files := listFile(root)
	for _, file := range files {
		if file.IsDir() {
			subDir := path.Join(root, file.Name())
			wg.Add(1)
			go walkDir(subDir, fileSize, wg)
		} else {
			fileSize <- file.Size()
		}
	}
}

func printInfo(nFiles, fileSize int64) {
	unit := "B"
	var fileSizeFloat float64
	const (
		_1K int64 = 1 << 10
		_1M int64 = 1 << 20
		_1G int64 = 1 << 30
	)
	if fileSize > _1G {
		fileSizeFloat = float64(fileSize) / float64(_1G)
		unit = "GB"
	} else if fileSize > _1M {
		fileSizeFloat = float64(fileSize) / float64(_1M)
		unit = "MB"
	} else if fileSize > _1K {
		fileSizeFloat = float64(fileSize) / float64(_1K)
		unit = "KB"
	}
	fmt.Printf("%d files\t%.2f %s\n", nFiles, fileSizeFloat, unit)
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		args = append(args, ".")
	}
	size := make(chan int64)
	var wg sync.WaitGroup
	for _, root := range args {
		wg.Add(1)
		go walkDir(root, size, &wg)
	}
	go func() {
		wg.Wait()
		close(size)
	}()

	var totalSize int64
	var nFiles int64
	for s := range size {
		nFiles++
		totalSize += s
	}
	printInfo(nFiles, totalSize)
}
