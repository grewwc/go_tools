package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	_1K int64 = 1 << 10
	_1M int64 = 1 << 20
	_1G int64 = 1 << 30
)

var lowerSizeBound float64 = -1

var threadControl = make(chan struct{}, 50)

var excludes = cw.NewConcurrentHashSet[string](nil, nil)

var types = cw.NewConcurrentHashSet[string](nil, nil)

var verbose = false

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

func walkDir(root string, fileInfoChan chan<- *cw.Tuple, wg *sync.WaitGroup) {
	defer wg.Done()
	files, err := listFile(root)
	if err != nil {
		return
	}
	for _, file := range files {
		if file.IsDir() {
			subDir := path.Join(root, file.Name())
			wg.Add(1)
			go walkDir(subDir, fileInfoChan, wg)
		} else {
			if !valid(file.Name()) {
				continue
			}
			fileInfo, err := file.Info()
			if err != nil {
				return
			}
			fileInfoChan <- cw.NewTuple(root, fileInfo)
		}
	}
}

func printInfo(nFiles, fileSize int64, numSpace int) {
	fs := formatFileSize(fileSize)
	fmt.Printf("%s%d files\t%s\n", strings.Repeat(" ", numSpace), nFiles, fs)
}

func formatFileSize(fileSize int64) string {
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
	} else {
		fileSizeFloat = float64(fileSize)
		unit = "B"
	}
	if fileSizeFloat < 1e-5 {
		return "0 B"
	}
	return fmt.Sprintf("%.3f %s", fileSizeFloat, unit)
}

func checkOneDirectory(root string) {
	if utilsw.IsRegular(root) && !valid(root) {
		return
	}
	var wg sync.WaitGroup
	wg.Add(1)
	fileInfoCh := make(chan *cw.Tuple)
	go walkDir(root, fileInfoCh, &wg)
	go func() {
		wg.Wait()
		close(fileInfoCh)
	}()

	var totalSize int64
	var nFiles int64
	subFiles := cw.NewLinkedList[*cw.Tuple]()
	for t := range fileInfoCh {
		s := t.Get(1).(os.FileInfo)
		nFiles++
		totalSize += s.Size()
		subFiles.PushBack(t)
	}
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}
	if (lowerSizeBound < 0) || (lowerSizeBound > 0 && totalSize >= int64(lowerSizeBound)) {
		fmt.Println(color.HiBlueString("%s", root))
		printInfo(nFiles, totalSize, 4)
	}
}

func getOnlyDirectories(root string) []string {
	return utilsw.LsDir(root,
		func(filename string) bool { return utilsw.IsDir(filename) },
		func(filename string) string { return filepath.Join(root, filename) })
}

func getOnlyFiles(root string) []string {
	return utilsw.LsDir(root,
		func(file string) bool { return !utilsw.IsDir(file) },
		func(filename string) string {
			return filepath.Join(root, filename)
		})
}

func getDirAndFiles(root string) []string {
	return utilsw.LsDir(root,
		func(filename string) bool { return !excludes.Contains(filename) },
		func(filename string) string { return filepath.Join(root, filename) })
}

func parseSize(size string) float64 {
	if len(size) == 0 {
		return -1
	}
	val, err := strconv.Atoi(size)
	if err == nil {
		return float64(val)
	}
	size, unit := size[:len(size)-1], string(size[len(size)-1])
	sizeFloat, err := strconv.ParseFloat(size, 64)
	if err != nil {
		log.Fatalln(err)
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

func checkOneFile(f string) {
	if !valid(f) {
		return
	}
	info, err := os.Stat(f)
	if err != nil {
		log.Println(color.RedString("failed get file info: %s", f))
		return
	}
	sz := info.Size()
	if sz > int64(lowerSizeBound) {
		fs := formatFileSize(sz)
		fmt.Printf("%s  \t%s\n", color.YellowString(f), fs)
	}
}

func check(f string) {
	if utilsw.IsDir(f) {
		checkOneDirectory(f)
	} else if utilsw.IsRegular(f) {
		checkOneFile(f)
	} else {
		log.Println(color.RedString("unknow file: %s", f))
	}
}

func getFirstDir(parsed *terminalw.Parser) string {
	if parsed.Empty() {
		return "."
	}
	args := parsed.Positional.ToStringSlice()
	if len(args) == 0 {
		return "."
	}
	return args[0]
}

func getExcludeFiles(parsed *terminalw.Parser) {
	ex := parsed.GetFlagValueDefault("ex", "")
	for _, file := range strw.SplitNoEmpty(ex, ",") {
		file = strings.Trim(file, " ")
		excludes.Add(file)
	}
}

func getTypes(parsed *terminalw.Parser) {
	t := parsed.GetFlagValueDefault("t", "")
	for _, file := range strw.SplitNoEmpty(t, ",") {
		file = strings.Trim(file, " ")
		if !strings.HasPrefix(file, ".") {
			file = "." + file
		}
		types.Add(file)
	}
}

func valid(file string) bool {
	notExcluded := !excludes.Contains(file)
	if !types.Empty() {
		return notExcluded && types.Contains(filepath.Ext(file))
	}
	return notExcluded
}

func main() {
	parser := terminalw.NewParser()
	parser.Bool("v", false, "list directries seperately")
	parser.Bool("d", false, "only list directries")
	parser.Bool("f", false, "only list regular files")
	parser.String("gt", "", "size greater than. (1.3g, 1m, 1K)")
	parser.String("t", "", "file types (e.g.: '.txt, pdf')")
	parser.String("ex", "", "exclude files or dirs (including subdirs having same name)")
	parser.Bool("h", false, "print help info")
	parser.ParseArgsCmd("v", "d", "f", "h")

	args := make([]string, 0)

	onlyFile := false
	onlyDir := false

	if parser.Empty() {
		args = append(args, ".")
		goto check
	} else if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}

	onlyFile = parser.ContainsAllFlagStrict("f")
	onlyDir = parser.ContainsAllFlagStrict("d")
	verbose = parser.ContainsAllFlagStrict("v") || onlyDir || onlyFile
	getExcludeFiles(parser)
	getTypes(parser)

	if verbose {
		rootDir := getFirstDir(parser)
		if onlyDir {
			args = append(args, getOnlyDirectories(rootDir)...)
		} else if onlyFile {
			args = append(args, getOnlyFiles(rootDir)...)
		} else {
			args = append(args, getDirAndFiles(rootDir)...)
		}
	}
	if parser.ContainsFlagStrict("gt") {
		lowerSizeBound = parseSize(parser.GetFlagValueDefault("gt", "10000g"))
	}
	if len(parser.Positional.ToStringSlice()) == 0 && !verbose {
		args = append(args, ".")
	}
	for _, file := range parser.Positional.ToStringSlice() {
		if utilsw.IsDir(file) && !onlyFile && valid(file) {
			args = append(args, file)
		} else if utilsw.IsRegular(file) && !onlyDir && valid(file) {
			args = append(args, file)
		} else {
			args = append(args, file)
		}
	}

check:
	for _, file := range args {
		check(file)
	}
}
