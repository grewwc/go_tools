package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	_1K int64 = 1 << 10
	_1M int64 = 1 << 20
	_1G int64 = 1 << 30
)

var lowerSizeBound float64 = -1

var threadControl = make(chan struct{}, 50)

func listFile(path string) ([]os.DirEntry, error) {
	threadControl <- struct{}{}
	defer func() { <-threadControl }()
	fileInfos, err := os.ReadDir(path)
	if err != nil {
		// fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	return fileInfos, nil
}

func walkDir(root string, fileSize chan<- int64, wg *sync.WaitGroup) {
	defer wg.Done()
	files, err := listFile(root)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			subDir := path.Join(root, file.Name())
			wg.Add(1)
			go walkDir(subDir, fileSize, wg)
		} else {
			fileInfo, err := file.Info()
			if err != nil {
				return
			}
			fileSize <- fileInfo.Size()
		}
	}
}

func printInfo(nFiles, fileSize int64, numSpace int) {
	unit := "B"
	var fileSizeFloat float64

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
	fmt.Printf("%s%d files\t%.2f %s\n", strings.Repeat(" ", numSpace), nFiles, fileSizeFloat, unit)
}

func checkOneDirectory(root string) {
	var wg sync.WaitGroup
	wg.Add(1)
	filesizeChan := make(chan int64)
	go walkDir(root, filesizeChan, &wg)
	go func() {
		wg.Wait()
		close(filesizeChan)
	}()

	var totalSize int64
	var nFiles int64
	for s := range filesizeChan {
		nFiles++
		totalSize += s
	}
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}

	if totalSize > int64(lowerSizeBound) {
		fmt.Printf(color.HiBlueString("%s\n", root))
		printInfo(nFiles, totalSize, 4)
	}
}

func getOnlyDirectories(root string) []string {
	result := make([]string, 0)
	for _, file := range utilsW.LsDir(root) {
		if !utilsW.IsDir(file) {
			continue
		}
		result = append(result, file)
	}
	return result
}

func parseSize(size string) float64 {
	if len(size) == 0 {
		return -1
	}
	size, unit := size[:len(size)-1], string(size[len(size)-1])
	sizeFloat, err := strconv.ParseFloat(size, 64)
	if err != nil {
		size, unit = size+unit, ""
	}
	unit = strings.ToLower(unit)
	switch unit {
	case "k":
		sizeFloat *= float64(_1K)
	case "m":
		sizeFloat *= float64(_1M)
	case "g":
		sizeFloat *= float64(_1G)
	}
	return sizeFloat
}

func main() {
	fs := flag.NewFlagSet("fs", flag.ExitOnError)
	fs.Bool("v", false, "list directries seperately")
	fs.String("gt", "", "size greater than. (1.3g, 1m, 1K)")
	parsed := terminalW.ParseArgsCmd("v")
	args := make([]string, 0)
	if parsed == nil {
		args = append(args, ".")
		goto check
	} else if parsed.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		return
	}
	if len(parsed.Positional.ToSlice()) == 0 {
		if parsed.ContainsFlagStrict("v") {
			for _, dir := range getOnlyDirectories(".") {
				args = append(args, dir)
			}
		}
	}
	if parsed.ContainsFlagStrict("gt") {
		lowerSizeBound = parseSize(parsed.GetFlagValueDefault("gt", "10000g"))
	}
	for _, dir := range parsed.Positional.ToStringSlice() {
		if utilsW.IsDir(dir) {
			args = append(args, dir)
		}
	}

check:
	for _, dir := range args {
		checkOneDirectory(dir)
	}
}
