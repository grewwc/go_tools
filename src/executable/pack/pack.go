package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

var (
	listOnly = false
)

var (
	excludeFileExtension = cw.NewTrie()
	fileExtension        = cw.NewTrie()
	excludeFileDir       = cw.NewLinkedList[string]()
)

// 控制打开文件数量
var ch = make(chan struct{}, 50)

func processTarGzFile(fname string, prefix string) {
	if !listOnly {
		fmt.Fprintf(color.Output, "untar \"%s\" to \"%s\"\n", fname, prefix)
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

		if err != nil && err != io.EOF {
			log.Println(err)
			os.Exit(1)
		}

		if header == nil {
			break
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

		if err == io.EOF {
			break
		}
	}
}

func clean(fname string) {
	fmt.Printf("cleaning %s\n", fname)
	if utilsw.IsExist(fname) {
		msg := fmt.Sprintf("error occurred, clean %q\n", fname)
		log.Fatalln(color.RedString(msg))
		os.Remove(fname)
	}
}

func printHelp(parser *terminalw.Parser) {
	parser.PrintDefaults()
	fmt.Printf("%s dest.tar.gz source_dir\n", utilsw.BaseNoExt(utilsw.GetCurrentFileName()))
}

func hasPrefix(dir string) bool {
	for ex := range excludeFileDir.Iter().Iterate() {
		d := ex.Value()
		if strings.HasPrefix(dir, d) {
			return true
		}
	}
	return false
}

func main() {
	parser := terminalw.NewParser()
	parser.String("ex", "", "exclude file/directory")
	parser.String("exclude", "", "exclude file/directory, (i.e.: ${somedir}/.git, NOT .git")
	parser.Bool("v", false, "verbose")
	parser.Bool("u", false, "untar. (e.g: untar src.tar.gz dest_directory)")
	parser.Bool("h", false, "print help info")
	parser.Bool("clean", true, "clean the zipped file if error occurs")
	parser.Bool("l", false, "only list files in the tar.gz")
	parser.String("nt", "", "exclude file type. separated by comma, dot is NOT required.")
	parser.String("t", "", "only include file type, if set, ignore -nt & -ex. separated by comma, dot is NOT required.")
	parser.Bool("prog", true, "show progress (default is verbose)")
	parser.ParseArgsCmd("v", "u", "h", "clean", "l", "prog")
	// fmt.Println(parser)
	if parser.Empty() || parser.ContainsFlagStrict("h") {
		printHelp(parser)
		return
	}

	nt := parser.GetFlagValueDefault("nt", "")
	nt = strings.ReplaceAll(nt, ",", " ")
	t := parser.GetFlagValueDefault("t", "")
	t = strings.ReplaceAll(t, ",", " ")
	for _, val := range strw.SplitNoEmpty(nt, " ") {
		if val[0] != '.' {
			val = "." + val
		}
		// fmt.Println("here", val)
		excludeFileExtension.Insert(val)
	}
	for _, val := range strw.SplitNoEmpty(t, " ") {
		if val[0] != '.' {
			val = "." + val
		}
		fileExtension.Insert(val)
	}

	// create tar files
	exclude, err := parser.GetFlagVal("ex")
	if err != nil || exclude == "" {
		exclude, _ = parser.GetFlagVal("exclude")
		err = nil
	}
	if exclude != "" {
		for _, ex := range strw.SplitByStrKeepQuotes(exclude, ",", `"`, false) {
			ex = strings.TrimSpace(ex)
			ex = strw.Reverse(ex) + string(filepath.Separator)
			excludeFileDir.PushBack(ex)
		}
	}

	verbose := parser.ContainsFlagStrict("v")
	showProgress := parser.MustGetFlagVal("prog")

	// fmt.Println("excludeset", excludeSet)
	args := parser.Positional.ToStringSlice()
	// fmt.Println("here", args)
	srcNames := []string{}
	var srcName string
	if len(args) < 1 {
		printHelp(parser)
		return
	}
	outName := args[0]

	if !strw.AnyEquals(filepath.Ext(outName), ".gz", ".tgz") {
		msg := color.RedString(fmt.Sprintf("%q is not a valid outname", outName))
		panic(msg)
	}

	if parser.ContainsFlagStrict("l") {
		listOnly = true
	}
	// extract tar files
	if parser.ContainsFlagStrict("u") || listOnly {
		if parser.ContainsFlagStrict("u") {
			fmt.Println(color.GreenString("e.g: untar src.tar.gz dest_directory"))
		}

		args := parser.Positional.ToStringSlice()
		var src, prefix string

		src = args[0]

		if !parser.ContainsFlagStrict("l") {
			if len(args) < 2 {
				printHelp(parser)
			}
			prefix = args[1]
		}
		processTarGzFile(src, prefix)
		os.Exit(0)
	}
	// to tar files
	if utilsw.IsExist(outName) {
		ans := utilsw.PromptYesOrNo(fmt.Sprintf("%s exists, overwrite? (y/n) ", color.HiRedString(outName)))
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
		printHelp(parser)
		return
	} else {
		srcName = args[1]
	}

	if srcName != "" {
		srcNames, err = filepath.Glob(srcName)
	}

	if err != nil {
		if parser.ContainsFlagStrict("clean") {
			clean(outName)
		}
		log.Fatalln(err)
	}

	allFiles := []string{}
	for _, srcName := range srcNames {
		q := cw.NewQueue[string]()
		q.Enqueue(srcName)
		for !q.Empty() {
			curr := q.Dequeue()
			ext := filepath.Ext(curr)

			if hasPrefix(strw.Reverse(curr)) ||
				(ext != "" && excludeFileExtension.Contains(ext)) ||
				(!fileExtension.Empty() && !fileExtension.Contains(ext)) {

				if verbose {
					fmt.Println("exclude: ", color.YellowString(curr))
				}
				continue
			}
			allFiles = append(allFiles, curr)
			subDirs := utilsw.LsDir(curr, nil, func(filename string) string {
				return filepath.Join(curr, filename)
			})

			for _, subDir := range subDirs {
				q.Enqueue(subDir)
			}
		}
	}

	if len(allFiles) == 0 {
		fmt.Println(color.RedString(fmt.Sprintf("%q don't contain any files\n", srcName)))
		if parser.ContainsFlagStrict("clean") {
			clean(outName)
		}
		return
	}
	if err = utilsw.TarGz(outName, allFiles, verbose, showProgress == "true"); err != nil {
		if parser.ContainsFlagStrict("clean") {
			clean(outName)
		}
		log.Fatalln(err)
	}
}
