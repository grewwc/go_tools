package syncDir

import (
	"go_tools/src/stringsW/slices"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func isDir(fname string) bool {
	finfo, err := os.Stat(fname)
	if err != nil {
		return false
	}

	return finfo.IsDir()
}

func isExist(fname string) bool {
	_, err := os.Stat(fname)
	return os.IsExist(err) || err == nil
}

func shouldIgnoreDir(fname string) bool {
	return slices.Contains(
		parsedAttrsMap["ignore"], filepath.ToSlash(fname),
	)
}

func isRegular(fname string) bool {
	finfo, err := os.Stat(fname)
	if err != nil {
		return false
	}

	return finfo.Mode().IsRegular()
}

func lsDir(fname string) []string {
	infos, err := ioutil.ReadDir(fname)
	if err != nil {
		log.Fatal(err)
	}
	res := make([]string, len(infos))
	for i, info := range infos {
		res[i] = info.Name()
	}
	return res
}
