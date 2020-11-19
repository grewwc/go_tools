package utilsW

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func LsDir(fname string) []string {
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

func IsTextFile(filename string) bool {
	buf, _ := ioutil.ReadFile(filename)
	t := http.DetectContentType(buf)
	// fmt.Println("textfile", t)
	return strings.HasPrefix(t, "text")
}
