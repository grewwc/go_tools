package utilsw

import (
	"bufio"
	"os"
	"strings"
)

func GetLine() string {
	RunCmd("stty sane", nil)
	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	return strings.TrimRight(s.Text(), "\n")
}
