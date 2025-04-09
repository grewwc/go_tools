package utilw

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/grewwc/go_tools/src/conw"
)

var DefaultExtensions = conw.NewSet()

func init() {
	DefaultExtensions.AddAll(".py", ".cpp", ".js", ".txt", ".h", ".hpp", ".c",
		".tex", ".html", ".css", ".java", ".go", ".cc", ".htm", ".ts", ".xml",
		".php", ".sc", "")
}

// LsDir returns slices containing contents of a directory
// if fname is a file, not a directory, return empty slice
func LsDir(fname string, filter func(filename string) bool, postProcess func(filename string) string) []string {
	if !IsDir(fname) {
		return []string{}
	}
	infos, err := os.ReadDir(fname)
	if err != nil {
		log.Fatalln(err)
	}
	res := make([]string, 0, len(infos))
	for _, info := range infos {
		name := info.Name()
		if filter != nil && !filter(name) {
			continue
		}
		if postProcess != nil {
			name = postProcess(name)
		}
		res = append(res, name)
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
		// name = filepath.ToSlash(name)
		if !IsDir(name) {
			// fmt.Println("here", name)
			// res["./"] = append(res["./"], name)
			dirName := filepath.Dir(name) + "/"
			res[dirName] = append(res[dirName], filepath.Base(name))
		} else {
			res[name] = LsDir(name, nil, nil)
			// fmt.Println("here", name, res)
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

func IsExecutableOwner(fname string) bool {
	if IsDir(fname) {
		return false
	}
	finfo, err := os.Stat(fname)
	if err != nil {
		return false
	}
	return finfo.Mode()&0100 != 0
}

func IsExecutableGroup(fname string) bool {
	if IsDir(fname) {
		return false
	}
	finfo, err := os.Stat(fname)
	if err != nil {
		return false
	}
	return finfo.Mode()&0010 != 0
}

func IsExecutableOther(fname string) bool {
	if IsDir(fname) {
		return false
	}
	finfo, err := os.Stat(fname)
	if err != nil {
		return false
	}
	return finfo.Mode()&0001 != 0
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
	if os.IsNotExist(err) {
		return true
	}
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
	if idx < 0 {
		return filename
	}
	return filename[:idx]
}

func isTextFile(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		return false
	}
	buf := make([]byte, 256)
	_, err = f.Read(buf)
	if err != nil {
		return false
	}
	return IsText(buf)
}

func IsTextFile(filename string) bool {
	if !IsRegular(filename) {
		return false
	}
	ext := filepath.Ext(filename)

	if ext != "" && DefaultExtensions.Contains(ext) {
		return true
	}

	info, err := os.Lstat(filename)
	if err != nil {
		fmt.Println(err)
		return false
	}
	firstBit := info.Mode().String()[0]
	if firstBit != '-' {
		return false
	}
	return isTextFile(filename)
}

func IsText(data []byte) bool {
	t := http.DetectContentType(data)
	// fmt.Println(t)
	return strings.HasPrefix(t, "text")
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

// Abs returns absolute path , ignore error
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

func CopyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if !IsExist(filepath.Dir(dest)) {
		if err = os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return err
		}
	}
	out, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0444)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func ExpandUser(filename string) string {
	if len(filename) == 0 || filename[:2] != "~/" {
		return filename
	}
	home := os.Getenv("HOME")
	if home == "" {
		return filename
	}
	if home[len(home)-1] != '/' {
		home += "/"
	}
	return strings.Replace(filename, "~/", home, 1)
}

// LsRegexWd returns absolute path
func LsRegex(regex string) ([]string, error) {
	regex = ExpandUser(regex)
	var dir string
	var err error

	regex = strings.ReplaceAll(regex, ".*", "\x00")
	regex = strings.ReplaceAll(regex, ".+", "\x01")

	regex = strings.ReplaceAll(regex, ".", "\\.")
	regex = strings.ReplaceAll(regex, "?", ".?")
	regex = strings.ReplaceAll(regex, "+", ".+")
	regex = strings.ReplaceAll(regex, "*", ".*")

	regex = strings.ReplaceAll(regex, "\x00", ".*")
	regex = strings.ReplaceAll(regex, "\x01", ".+")

	idx := strings.LastIndexByte(regex, '/')
	if idx > 0 {
		dir = regex[:idx]
	} else {
		dir, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	regex = fmt.Sprintf("^%s$", regex)
	// fmt.Println("==>", regex)
	res := make([]string, 0)
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}

	for _, file := range LsDir(dir, nil, nil) {
		if re.MatchString(filepath.Join(dir, file)) {
			res = append(res, filepath.Join(dir, file))
		}
	}
	return res, nil
}

// ExpandWd only works for ./ prefix
func ExpandWd(filename string) string {
	if len(filename) == 0 || filepath.IsAbs(filename) {
		return filename
	}
	home, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if home == "" {
		return filename
	}
	if filename[:2] == "./" {
		if home[len(home)-1] != '/' {
			home += "/"
		}
		return strings.Replace(filename, "./", home, 1)
	} else {
		return filepath.Join(home, filename)
	}
}
