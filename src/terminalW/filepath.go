package terminalW

import (
	"path"
	"path/filepath"
)

func Glob(pattern, rootPath string) ([]string, error) {
	pattern = path.Join(rootPath, pattern)
	return filepath.Glob(pattern)
}
