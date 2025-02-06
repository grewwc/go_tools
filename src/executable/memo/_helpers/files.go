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

// LogMoveImages moves files or directories to a specified location and logs the operation details.
// This function receives the type of file and the absolute name of the file or directory as parameters,
// and returns a string recording the details of the move operation.
func LogMoveImages(type_, absName string) string {
	// Initialize the move log as an empty string
	moveLog := ""

	// Get all configuration information
	config := utilsW.GetAllConfig()

	// Get the root directory from the configuration, if not set, use the default value
	root := config.GetOrDefault(rootKey, "").(string)

	// Get the home directory environment variable
	home := os.Getenv("HOME")

	// If the root directory is empty, check and set based on the HOME environment variable
	if root == "" {
		if home == "" {
			// If HOME is also empty, throw a panic
			panic("$HOME is emptys")
		}
		// Set the root directory to the home directory plus the rootKey
		root = filepath.Join(home, rootKey)
	}

	// Initialize the file path list
	filepaths := make([]string, 0)

	// Resolve the absolute path of the file or directory
	absName = utilsW.ExpandWd(absName)
	absName = utilsW.ExpandUser(absName)

	// Check if the path is a directory
	if utilsW.IsDir(absName) {
		// If it's a directory, get all file paths in the directory
		for _, fname := range utilsW.LsDir(absName, nil) {
			filepaths = append(filepaths, filepath.Join(absName, fname))
		}
	} else {
		// If it's a single file, directly add it to the file path list
		filepaths = []string{absName}
	}

	// Initialize the retry count
	redoCnt := 0

	// Label for retrying file move operations
redo:
	for _, name := range filepaths {
		// Construct the target file path
		target := filepath.Join(root, type_, filepath.Base(name))

		// Build the log message for the move operation and add it to the log
		moveLog += buildLogMsg(name, target)
		moveLog += "\n"

		// Attempt to copy the file to the target location
		if err := utilsW.CopyFile(name, target); err != nil && redoCnt < 1 {
			// If the copy fails and the retry count is less than 1, prepare to retry

			// Get the list of files matching the regular expression in the current directory
			files, err := utilsW.LsRegex(absName)
			if err != nil {
				// If an error occurs during the regular expression match, throw a panic
				panic(err)
			}

			// Update the file path list to the result of the regular expression match
			filepaths = files

			// Increase the retry count
			redoCnt++

			// Jump back to the redo label to retry
			goto redo
		} else {
			// If the copy is successful or no retry is needed, attempt to delete the original file
			if err = os.Remove(name); err != nil {
				// If deletion fails, print an error message
				fmt.Printf("failed to remove the file: %s\n", name)
				fmt.Println(err)
			}
		}
	}

	// Return the move log
	return moveLog
}

func buildLogMsg(src, dest string) string {
	return fmt.Sprintf("%s -> %s", src, filepath.Dir(dest))
}
