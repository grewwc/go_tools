package utilsW

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
