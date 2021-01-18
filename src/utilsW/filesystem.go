package utilsW

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/grewwc/go_tools/src/containerW"
)

var DefaultExtensions = containerW.NewSet()

func init() {
	DefaultExtensions.AddAll(".py", ".cpp", ".js", ".txt", ".h", ".hpp", ".c",
		".tex", ".html", ".css", ".java", ".go", ".cc", ".htm", ".ts", ".xml",
		".php", ".sc", "")
}

// LsDir returns slices containing contents of a directory
// if fname is a file, not a directory, return empty slice
func LsDir(fname string) []string {
	if !IsDir(fname) {
		return []string{}
	}
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

func LsDirGlob(fname string) map[string][]string {
	names, err := filepath.Glob(fname)
	if err != nil {
		log.Println(err)
		return nil
	}
	res := make(map[string][]string)
	for _, name := range names {
		if !IsDir(name) {
			res["./"] = append(res["./"], name)
		} else {
			res[name] = LsDir(name)
		}
	}
	return res
}

func IsDir(fname string) bool {
	finfo, err := os.Stat(fname)
	if os.IsNotExist(err) || finfo == nil {
		return false
	}

	return finfo.IsDir()
}

func IsRegular(fname string) bool {
	finfo, err := os.Stat(fname)
	if os.IsNotExist(err) || finfo == nil {
		return false
	}

	return finfo.Mode().IsRegular()
}

func IsExist(fname string) bool {
	_, err := os.Stat(fname)
	return os.IsExist(err) || err == nil
}

func GetDirOfTheFile() string {
	_, dir, _, _ := runtime.Caller(1)
	return filepath.Dir(dir)
}

// IsNewer if filename1 newer than filename2
func IsNewer(filename1, filename2 string) bool {
	info1, err := os.Stat(filename1)
	if err != nil {
		log.Fatalln(err)
	}
	info2, err := os.Stat(filename2)
	if err != nil {
		log.Fatalln(err)
	}

	return info1.ModTime().After(info2.ModTime())
}

func GetCurrentFileNameAbs() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalln(ok)
	}

	return filename
}

func GetCurrentFileName() string {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		log.Fatalln(ok)
	}

	return filepath.Base(filename)
}

func TrimFileExt(filename string) string {
	idx := strings.LastIndex(filename, ".")
	return filename[:idx]
}

func isTextFile(filename string) bool {
	buf, _ := ioutil.ReadFile(filename)
	t := http.DetectContentType(buf)
	return strings.HasPrefix(t, "text")
}

func IsTextFile(filename string) bool {
	ext := filepath.Ext(filename)

	if ext != "" && DefaultExtensions.Contains(ext) {
		return true
	}

	info, err := os.Lstat(filename)
	if err != nil {
		Println(err)
		return false
	}
	firstBit := info.Mode().String()[0]
	if firstBit != '-' {
		return false
	}
	return isTextFile(filename)
}

func GetDirSize(dirname string) (int64, error) {
	var size int64
	err := filepath.Walk(dirname, func(prefix string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

// Abs ignore error
// return "" as representing error
func Abs(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, "~/") {
		userDir, err := os.UserHomeDir()
		if err != nil {
			log.Println(err)
			return ""
		}
		path = strings.ReplaceAll(path, "~", userDir)
	}
	path, err := filepath.Abs(path)
	if err != nil {
		log.Println(err)
		return ""
	}
	return path
}
