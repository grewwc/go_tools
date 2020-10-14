package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/grewwc/go_tools/src/stringsW/slices"
)

func trimSpace(strs []string) {
	for i := range strs {
		strs[i] = strings.TrimSpace(strs[i])
	}
}

func getFileUnixNano(fname string) int64 {
	finfo, err := os.Stat(fname)
	if err != nil {
		// log.Printf("cannot get mode time of %q\n", filepath.ToSlash(fname))
		return -1
	}
	return finfo.ModTime().UnixNano()
}

func clean(fname string) string {
	return filepath.ToSlash(filepath.Clean(fname))
}

func shouldIgnoreDir(fname string) bool {
	return slices.Contains(
		parsedAttrsMap["ignore"], filepath.ToSlash(fname),
	)
}
