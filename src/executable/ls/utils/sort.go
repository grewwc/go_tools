package _lsW

import (
	"sort"

	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	NewerFirst = iota
	OlderFirst
	Unsort
)

type sortByModifiedDate struct {
	files      []string
	newerFirst bool
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

func SortByModifiedDate(fileSlice []string, sortType int) []string {
	var newerFirst bool
	switch sortType {
	case NewerFirst:
		newerFirst = true
	case OlderFirst:
		newerFirst = false
	default:
		newerFirst = false
	}
	s := sortByModifiedDate{files: fileSlice, newerFirst: newerFirst}
	sort.Sort(s)
	return s.files
}
