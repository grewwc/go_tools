package utilsW

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TarGz add all srcNames to outName as tar.gz file
func TarGz(outName string, srcNames []string, verbose bool) error {
	out, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer out.Close()

	gw := gzip.NewWriter(out)
	tw := tar.NewWriter(gw)
	defer gw.Close()
	defer tw.Close()

	for _, filename := range srcNames {
		info, err := os.Stat(filename)
		if err != nil {
			return err
		}

		th, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}
		th.Name = filepath.ToSlash(filename)

		if err = tw.WriteHeader(th); err != nil {
			return err
		}
		if !IsDir(Abs(filename)) {
			src, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer src.Close()
			if _, err = io.Copy(tw, src); err != nil {
				return err
			}
		}
		if verbose {
			fmt.Println(filename)
		}
	}
	return nil
}

// ReadString read all content from fname
func ReadString(fname string) string {
	f, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return string(b)
}
