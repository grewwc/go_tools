package _helpers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	rootKey = "move_images"
)

func LogMoveImages(type_, absName string) string {
	moveLog := ""
	config := utilsW.GetAllConfig()
	root := config.GetOrDefault(rootKey, "").(string)
	home := os.Getenv("HOME")
	if root == "" {
		if home == "" {
			panic("$HOME is emptys")
		}
		root = filepath.Join(home, rootKey)
	}

	filepaths := make([]string, 0)
	absName = utilsW.ExpandWd(absName)
	absName = utilsW.ExpandUser(absName)
	if utilsW.IsDir(absName) {
		for _, fname := range utilsW.LsDir(absName) { // directory
			filepaths = append(filepaths, filepath.Join(absName, fname))
		}
	} else { // only 1 file
		filepaths = []string{absName}
	}
	redoCnt := 0
redo:
	for _, name := range filepaths {
		target := filepath.Join(root, type_, filepath.Base(name))
		moveLog += buildLogMsg(name, target)
		moveLog += "\n"
		// fmt.Println(name, target)
		if err := utilsW.CopyFile(name, target); err != nil && redoCnt < 1 {
			files, err := utilsW.LsRegex(absName)
			if err != nil {
				panic(err)
			}
			filepaths = files
			redoCnt++
			goto redo
		}
	}

	return moveLog
}

func buildLogMsg(src, dest string) string {
	return fmt.Sprintf("%s -> %s", src, filepath.Dir(dest))
}
