package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

var (
	listOnly = false
)

var (
	excludeFileExtension = containerW.NewTrie()
	fileExtension        = containerW.NewTrie()
)

// 控制打开文件数量
var ch = make(chan struct{}, 50)

func processTarGzFile(fname string, prefix string) {
	if !listOnly {
		fmt.Fprintf(color.Output, "untar \"%s\" to \"%s\"\n", color.GreenString(fname), color.YellowString(prefix))
	}
	ch <- struct{}{}
	defer func(ch <-chan struct{}) {
		<-ch
	}(ch)

	// first open the file
	f, err := os.Open(fname)
	if err != nil {
		return
	}
	defer f.Close()

	// open as gzip
	gf, err := gzip.NewReader(f)

	if err != nil {
		panic(err)
	}

	defer gf.Close()

	// open as tar
	tf := tar.NewReader(gf)

	for {
		header, err := tf.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if listOnly {
				continue
			}
			if err = os.MkdirAll(filepath.Join(prefix, header.Name), 0774); err != nil {
				log.Fatalln(err)
			}
		case tar.TypeReg:
			if listOnly {
				fmt.Println(header.Name)
				continue
			}
			outFile, err := os.Create(filepath.Join(prefix, header.Name))
			if err != nil {
				panic(err)
			}
			if _, err = io.Copy(outFile, tf); err != nil {
				panic(err)
			}
		default:
			panic(fmt.Sprintf("wrong,%v,%s", header.Typeflag, color.RedString(header.Name)))
		}
	}
}

func clean(fname string) {
	fmt.Printf("cleaning %s\n", fname)
	if utilsW.IsExist(fname) {
		msg := fmt.Sprintf("error occurred, clean %q\n", fname)
		log.Fatalln(color.RedString(msg))
		os.Remove(fname)
	}
}

func main() {
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.String("ex", "", "exclude file/directory")
	fs.String("exclude", "", "exclude file/directory, (i.e.: ${somedir}/.git, NOT .git")
	fs.Bool("v", false, "verbose")
	fs.Bool("u", false, "untar")
	fs.Bool("h", false, "print help info")
	fs.Bool("clean", true, "clean the zipped file if error occurs")
	fs.Bool("l", false, "only list files in the tar.gz")
	fs.String("nt", "", "exclude file type")
	fs.String("t", "", "only include file type, if set, ignore -nt & -ex")

	parsedResults := terminalW.ParseArgsCmd("v", "u", "h", "clean", "l")
	// fmt.Println(parsedResults)
	if parsedResults == nil || parsedResults.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		fmt.Println(color.GreenString("targo thesis.tar.gz thesis_folder"))
		return
	}

	nt := parsedResults.GetFlagValueDefault("nt", "")
	nt = strings.ReplaceAll(nt, ",", "")
	t := parsedResults.GetFlagValueDefault("t", "")
	t = strings.ReplaceAll(t, ",", "")
	for _, val := range stringsW.SplitNoEmpty(nt, " ") {
		if val[0] != '.' {
			val = "." + val
		}
		// fmt.Println("here", val)
		excludeFileExtension.Insert(val)
	}
	for _, val := range stringsW.SplitNoEmpty(t, " ") {
		if val[0] != '.' {
			val = "." + val
		}
		fileExtension.Insert(val)
	}

	// create tar files
	exclude, err := parsedResults.GetFlagVal("ex")
	if err != nil || exclude == "" {
		exclude, _ = parsedResults.GetFlagVal("exclude")
	}
	if exclude != "" {
		exclude = utilsW.Abs(exclude)
	}

	excludes, err := filepath.Glob(exclude)
	if err != nil {
		log.Println(err)
		return
	}

	verbose := parsedResults.ContainsFlagStrict("v")
	excludeSet := containerW.NewSet()

	for _, ex := range excludes {
		filepath.Walk(ex, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			excludeSet.Add(utilsW.Abs(path))
			return nil
		})
	}
	// fmt.Println("excludeset", excludeSet)
	args := parsedResults.Positional.ToStringSlice()
	// fmt.Println("here", args)
	srcNames := []string{}
	var srcName string
	outName := args[0]

	if filepath.Ext(outName) != ".gz" {
		msg := color.RedString(fmt.Sprintf("%q is not a valid outname", outName))
		panic(msg)
	}

	if parsedResults.ContainsFlagStrict("l") {
		listOnly = true
	}
	// extract tar files
	if parsedResults.ContainsFlagStrict("u") || listOnly {
		fmt.Println(color.GreenString("e.g: untar src.tar.gz dest_directory"))

		args := parsedResults.Positional.ToStringSlice()
		var src, prefix string

		src = args[0]

		if !parsedResults.ContainsFlagStrict("l") {
			prefix = args[1]
		}
		processTarGzFile(src, prefix)
		os.Exit(0)
	}
	// to tar files
	if utilsW.IsExist(outName) {
		ans := utilsW.PromptYesOrNo(fmt.Sprintf("%s exists, overwrite? (y/n) ", color.HiRedString(outName)))
		if ans {
			fmt.Printf("overrite %s!\n", color.RedString(outName))
		} else {
			fmt.Println("quit")
			return
		}
	}

	if len(args) > 2 {
		srcNames = args[1:]
	} else if len(args) <= 1 {
		fs.PrintDefaults()
		return
	} else {
		srcName = args[1]
	}

	if srcName != "" {
		srcNames, err = filepath.Glob(srcName)
	}

	if err != nil {
		if parsedResults.ContainsFlagStrict("clean") {
			clean(outName)
		}
		log.Fatalln(err)
	}

	allFiles := []string{}
	for _, srcName := range srcNames {
		filepath.Walk(srcName, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			absPath := utilsW.Abs(path)
			ext := filepath.Ext(path)
			if t != "" {
				// 没有文件后缀的也忽略
				if fileExtension.Search(ext) && ext != "" {
					allFiles = append(allFiles, path)
				}
			} else if !excludeSet.Contains(absPath) && (ext == "" || !excludeFileExtension.Search(ext)) {
				allFiles = append(allFiles, path)
			} else if verbose {
				fmt.Println("exclude: ", color.YellowString(path))
			}
			return nil
		})
	}

	if len(allFiles) == 0 {
		fmt.Println(color.RedString(fmt.Sprintf("%q don't contain any files\n", srcName)))
		if parsedResults.ContainsFlagStrict("clean") {
			clean(outName)
		}
		return
	}
	if err = utilsW.TarGz(outName, allFiles, verbose); err != nil {
		if parsedResults.ContainsFlagStrict("clean") {
			clean(outName)
		}
		log.Fatalln(err)
	}
	fmt.Println()
}
