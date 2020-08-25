package systools

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Fatal(err)
		}
	}
	return true
}

//Move ONLY files in one directory to another directory
//"from" and "to" are both directory Absolute path.
func CopyFile(to string, from string) {

	//change "." to working directory
	if strings.HasPrefix(from, ".") {
		pathFromDot, err := filepath.Abs(".")
		if err != nil {
			log.Fatal(err)

		}
		from = filepath.Join(pathFromDot, filepath.Base(from))
	}

	if strings.HasPrefix(to, ".") {
		pathFromDot, err := filepath.Abs(".")
		if err != nil {
			log.Fatal(err)
		}
		to = filepath.Join(pathFromDot, filepath.Base(to))
	}
	fmt.Println(IsExist(to))
	if !IsExist(to) {
		if err := os.MkdirAll(to, os.ModeDir|os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}
	allFiles, err := filepath.Glob(filepath.Join(from, "*"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(to, from)
	for _, f := range allFiles {
		stat, err := os.Stat(f)
		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
		if !stat.IsDir() {
			src, err := os.Open(f)
			if err != nil {
				log.Fatal(err)
			}
			dst, err := os.Create(filepath.Join(to, filepath.Base(f)))
			if err != nil {
				log.Fatal(err)
			}
			defer dst.Close()
			_, err = io.Copy(dst, src)
			if err != nil && err != io.EOF {
				log.Fatal(err)
			}
			defer src.Close()

		}
	}

}
