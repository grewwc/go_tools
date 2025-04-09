//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/utilw"
	"github.com/grewwc/go_tools/src/windowsW"
)

func init() {
	windowsW.EnableVirtualTerminal()
}

func getCreateTime(filename string) (time.Time, error) {
	finfo, err := os.Stat(filename)
	if err != nil {
		return time.Now(), err
	}
	data := finfo.Sys().(*syscall.Win32FileAttributeData)
	return time.Unix(0, data.CreationTime.Nanoseconds()), nil
}

func modeStrToNum(mode string) string {
	owner := mode[1:4]
	group := mode[4:7]
	other := mode[7:]

	m := map[byte]int{
		'r': 4,
		'w': 2,
		'x': 1,
		'-': 0,
	}
	ownerVal, groupVal, otherVal := 0, 0, 0
	for i := range owner {
		ownerVal += m[owner[i]]
		groupVal += m[group[i]]
		otherVal += m[other[i]]
	}
	return fmt.Sprintf("0%d%d%d", ownerVal, groupVal, otherVal)
}

func processSingle(filename string) {
	f, err := os.Stat(filename)
	if err != nil {
		log.Fatalln(err)
	}
	cTime, err := getCreateTime(filename)
	if err != nil {
		log.Fatalln(err)
	}
	mTime := f.ModTime()
	size, err := utilw.GetDirSize(filename)
	if err != nil {
		log.Fatalln(err)
	}
	modeStr := f.Mode().String()
	modeNum := modeStrToNum(modeStr)

	fmt.Printf("    File: %s\tSize: %s\tAccess: (%s/%s)\n",
		color.HiGreenString(filename), strw.FormatInt64(size), modeNum, modeStr)
	fmt.Printf("  Create: %v\n", cTime)
	fmt.Printf("  Modify: %v\n", mTime)
}

func main() {
	args := os.Args[1:]
	for _, filename := range args {
		files := utilw.LsDirGlob(filename)
		for d, fnames := range files {
			if d == "./" {
				for _, fname := range fnames {
					processSingle(fname)
				}
			} else {
				processSingle(d)
			}
		}
	}
}
