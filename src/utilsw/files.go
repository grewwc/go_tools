package utilsw

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/grewwc/go_tools/src/strw"
)

// TarGz add all srcNames to outName as tar.gz file
func TarGz(outName string, srcNames []string, verbose, showProgress bool) error {
	out, err := os.Create(outName)
	if err != nil {
		return err
	}
	defer out.Close()
	cnt := 0

	gw := gzip.NewWriter(out)
	tw := tar.NewWriter(gw)
	defer gw.Close()
	defer tw.Close()

	var buf [8 * 1024 * 1024]byte

	for _, filename := range srcNames {
		info, err := os.Stat(filename)
		if err != nil {
			return err
		}
		link := info.Name()
		// read link
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			link, err = os.Readlink(filename)
			if err != nil {
				log.Fatalln(err)
			}
		}

		th, err := tar.FileInfoHeader(info, link)
		if err != nil {
			return err
		}
		th.Name = filepath.ToSlash(filename)

		if err = tw.WriteHeader(th); err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			src, err := os.Open(filename)
			if err != nil {
				return err
			}
			if _, err = io.CopyBuffer(tw, src, buf[:]); err != nil {
				src.Close()
				return err
			}
			src.Close()
			if verbose {
				msg := filename
				if showProgress {
					cnt++
					msg = fmt.Sprintf("%s (%d)", msg, cnt)
				}
				fmt.Println(msg)
			}
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
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// WriteToFile will clean the original content!!
func WriteToFile(filename string, buf []byte) error {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0664)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, err = f.Write(buf)
	return err
}

func InputWithEditor(originalContent string, vs bool) (res string) {
	fname := uuid.New().String() + ".txt"
	var cmd *exec.Cmd
	f, err := os.OpenFile(fname, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	f.WriteString(originalContent)
	f.Close()

	switch GetPlatform() {
	case MAC, LINUX:
		if !vs {
			cmd = exec.Command("vim", fname)
		} else {
			cmd = exec.Command("code", "-w", fname)
		}
	case WINDOWS:
		cmd = exec.Command("cmd.exe", "/C", "notepad.exe", fname)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Run()
	res = ReadString(fname)

	t := time.After(time.Second)
	ch := make(chan interface{})
	go func() {
		for err := os.Remove(fname); err != nil; err = os.Remove(fname) {
		}
		close(ch)
	}()

	select {
	case <-ch:
		return
	case <-t:
		return
	}
}

func GetFileMode(fname string) os.FileMode {
	info, err := os.Stat(fname)
	if err != nil {
		panic(err)
	}
	return info.Mode()
}

func GetFileSize(fname string) int64 {
	if IsDir(fname) {
		fmt.Printf("%s is a directory!", fname)
		return 0
	}
	info, err := os.Stat(fname)
	if err != nil {
		panic(err)
	}
	return info.Size()
}

func BaseNoExt(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)
	return strw.StripSuffix(base, ext)
}
