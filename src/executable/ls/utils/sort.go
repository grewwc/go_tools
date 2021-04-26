package _lsW

import (
	"path/filepath"
	"sort"
	"strconv"

	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	NewerFirst = iota
	OlderFirst
	NumberSmallerFirst
	NumberLargestFirst
	Unsort
)

func getNum(s string) int {
	var chars []rune
	for _, char := range s {
		if (char >= '0') && (char <= '9') {
			chars = append(chars, char)
		}
	}
	if len(chars) > 0 {
		res, _ := strconv.Atoi(string(chars))
		return res
	}
	return -1
}

type sortByModifiedDate struct {
	rootDir    string
	files      []string
	newerFirst bool
}

func NewSortByModifiedDate(rootDir string, files []string, newerFirst bool) *sortByModifiedDate {
	res := sortByModifiedDate{rootDir: rootDir, newerFirst: newerFirst}
	absFileSlice := make([]string, len(files))
	for i, f := range files {
		absFileSlice[i] = filepath.Join(rootDir, f)
	}
	res.files = absFileSlice
	return &res
}

func (this sortByModifiedDate) Len() int {
	return len(this.files)
}

func (this sortByModifiedDate) Swap(i, j int) {
	this.files[i], this.files[j] = this.files[j], this.files[i]
}

func (this sortByModifiedDate) Less(i, j int) bool {
	if this.newerFirst {
		return utilsW.IsNewer(this.files[i], this.files[j])
	}
	return utilsW.IsNewer(this.files[j], this.files[i])
}

type sortByStringNum struct {
	files        []string
	smallerFirst bool
}

func (this sortByStringNum) Len() int {
	return len(this.files)
}

func (this sortByStringNum) Swap(i, j int) {
	this.files[i], this.files[j] = this.files[j], this.files[i]
}

func (this sortByStringNum) Less(i, j int) bool {
	if this.smallerFirst {
		return getNum(this.files[i]) < getNum(this.files[j])
	}
	return getNum(this.files[i]) > getNum(this.files[j])
}

func SortByModifiedDate(rootDir string, fileSlice []string, sortType int) []string {
	var newerFirst bool
	switch sortType {
	case NewerFirst:
		newerFirst = true
	case OlderFirst:
		newerFirst = false
	default:
		newerFirst = false
	}
	// construct absolute path because need to use os.stat to check file modified time
	s := NewSortByModifiedDate(rootDir, fileSlice, newerFirst)
	sort.Sort(s)
	// need to return relative paths
	res := make([]string, len(s.files))
	for i, file := range s.files {
		res[i] = filepath.Base(file)
	}
	return res
}

func SortByStringNum(fileSlice []string, sortType int) []string {
	var smallerFirst bool
	if sortType == NumberSmallerFirst {
		smallerFirst = true
	} else {
		smallerFirst = false
	}
	s := sortByStringNum{files: fileSlice, smallerFirst: smallerFirst}
	sort.Sort(s)
	return s.files
}
