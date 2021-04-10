package main

import (
	"archive/tar"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

func processTarGzFile(fname string, prefix string) {
	fmt.Printf("untar %q to %q\n", fname, prefix)
	// first open the file
	f, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	// open as gzip
	gf, err := gzip.NewReader(f)

	if err != nil {
		fmt.Println(err)
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
			os.MkdirAll(filepath.Join(prefix, header.Name), os.ModeDir)
		case tar.TypeReg:
			buf := make([]byte, int(header.Size))
			_, err = tf.Read(buf)
			if err != io.EOF && err != nil {
				log.Println(err)
				os.Exit(1)
			}
			ioutil.WriteFile(filepath.Join(prefix, header.Name), buf, 0755)
		}
	}
}

func clean(fname string) {
	fmt.Printf("cleaning %s\n", fname)
	if utilsW.IsExist(fname) {
		fmt.Printf("error occurred, clean %q\n", fname)
		os.Remove(fname)
	}
}

func main() {
	fs := flag.NewFlagSet("parser", flag.ExitOnError)
	fs.String("ex", "", "exclude file/directory")
	fs.String("exclude", "", "exclude file/directory")
	fs.Bool("v", false, "verbose")
	fs.Bool("u", false, "untar")
	fs.Bool("h", false, "print help info")
	fs.Bool("clean", false, "if clean the zipped file if error occurs")

	parsedResults := terminalW.ParseArgsCmd("v", "u", "h", "clean")
	if parsedResults == nil || parsedResults.ContainsFlagStrict("h") {
		fs.PrintDefaults()
		fmt.Println("targo thesis.tar.gz thesis_folder")
		return
	}

	// extract tar files
	if parsedResults.ContainsFlag("u") {
		fmt.Println("e.g: untar src.tar.gz dest_directory")

		args := parsedResults.Positional.ToStringSlice()
		// fmt.Println(args)
		if len(args) != 2 {
			fmt.Println("need 2 arguments as hinted above")
			os.Exit(1)
		}
		src := args[0]
		prefix := args[1]

		processTarGzFile(src, prefix)
		os.Exit(0)
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
			if info.IsDir() {
				return nil
			}
			if err != nil {
				return err
			}
			excludeSet.Add(path)
			return nil
		})
	}

	args := parsedResults.Positional.ToStringSlice()
	srcNames := []string{}
	var srcName string
	outName := args[0]
	if len(args) > 2 {
		srcNames = args[1:]
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
			// abspath := utilsW.Abs(path)
			if !excludeSet.Contains(path) {
				allFiles = append(allFiles, path)
				if verbose {
					fmt.Println(path)
				}
			} else if verbose {
				fmt.Println("exclude: ", path)
			}
			return nil
		})
	}

	if len(allFiles) == 0 {
		fmt.Printf("%q don't contain any files\n", srcName)
		if parsedResults.ContainsFlagStrict("clean") {
			clean(outName)
		}
		return
	}
	if err = utilsW.TarGz(outName, allFiles); err != nil {
		if parsedResults.ContainsFlagStrict("clean") {
			clean(outName)
		}
		log.Fatalln(err)
	}
	fmt.Println()
}
