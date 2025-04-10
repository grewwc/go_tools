package utilw

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/atotto/clipboard"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/peterh/liner"
	"github.com/petermattis/goid"
)

func toString(numTab int, obj interface{}, ignoresFieldName ...string) string {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct {
		return fmt.Sprintf("%v", obj)
	}
	copyV := reflect.New(v.Type()).Elem()
	copyV.Set(v)
	structName := fmt.Sprintf("%v {", t)
	s := cw.NewSet()
	for _, ignore := range ignoresFieldName {
		s.Add(ignore)
	}
	first := true
	buf := bytes.NewBufferString(structName)
	for i := 0; i < t.NumField(); i++ {
		fieldName := t.Field(i).Name
		if s.Contains(fieldName) {
			continue
		}
		if !first {
			buf.WriteString(strings.Repeat(" ", len(structName)+1+numTab))
		} else {
			first = false
			buf.WriteString(" ")
		}
		buf.WriteString(fieldName)
		buf.WriteString(": ")
		var val string
		fieldVal := copyV.Field(i)
		fieldVal = reflect.NewAt(fieldVal.Type(), unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		if fieldVal.Type() == reflect.TypeOf(time.Time{}) {
			val = (fieldVal.Interface().(time.Time)).Local().Format("2006-01-02/15:04:05")
		} else {
			// val = fmt.Sprintf("%v", v.Field(i))
			val = toString(len(structName)+len(fieldName)+3, fieldVal.Interface())
		}
		buf.WriteString(val)
		buf.WriteString("\n")
	}
	buf.WriteString(strings.Repeat(" ", numTab))
	buf.WriteString("}")
	return buf.String()
}

func ToString(obj interface{}, ignoresFieldName ...string) string {
	return toString(0, obj, ignoresFieldName...)
}

func OpenUrlInBrowswer(url string) {
	var cmdStr string
	var args []string

	switch GetPlatform() {
	case WINDOWS:
		cmdStr = "cmd"
		args = []string{"/C", "start", "", url}
	case LINUX:
		cmdStr = "xdg-open"
		args = []string{url}
	case MAC:
		cmdStr = "open"
		args = []string{url}
	}
	cmd := exec.Command(cmdStr, args...)
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func TimeoutWait(wg *sync.WaitGroup, timeout time.Duration) error {
	c := make(chan interface{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

func GetTerminalSize() (h, w int, err error) {
	var cmd *exec.Cmd
	if GetPlatform() == WINDOWS {
		// cmd = exec.Command("powershell", "-command", "&{$H=get-host;$H.ui.rawui.WindowSize;}")
		cmd = exec.Command("sh", "-c", "/bin/stty size")
	} else {
		cmd = exec.Command("/bin/stty", "size")
	}
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return
	}
	size := strw.SplitNoEmpty(strings.TrimSpace(string(out)), " ")
	h, err = strconv.Atoi(size[0])
	if err != nil {
		return
	}
	w, err = strconv.Atoi(size[1])
	if err != nil {
		return
	}
	return
}

func PromptYesOrNo(msg string) bool {
	if len(msg) == 0 || msg[len(msg)-1] != ' ' {
		msg += " "
	}
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	ans := strings.TrimSpace(scanner.Text())
	return strings.ToLower(ans) == "y"
}

func UserInput(msg string, multiline bool) string {
	defer RunCmd("stty sane", os.Stdin)
	line := liner.NewLiner()
	line.SetMultiLineMode(multiline)
	defer line.Close()
	historyFile := ExpandUser("~/.liner_histroy")
	f, err := os.OpenFile(historyFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer line.Close()
	line.ReadHistory(f)
	line.SetCtrlCAborts(true)

	var lines []string
	for {
		input, err := line.Prompt(msg)
		if err != nil {
			if err == liner.ErrPromptAborted {
				fmt.Println("Exit.")
				os.Exit(0)
			}
			if err != io.EOF {
				panic(err)
			}
		}

		if !multiline {
			if input != "" {
				line.AppendHistory(input)
			}
			if _, err := line.WriteHistory(f); err != nil {
				panic(err)
			}

			return input
		}
		lines = append(lines, input)
		if err == io.EOF {
			break
		}
	}
	return strings.Join(lines, "\n")
}

// GetCommandList seperate cmd to []string
func GetCommandList(cmd string) []string {
	cmd = strings.ReplaceAll(cmd, ",", " ")
	return strw.SplitNoEmpty(cmd, " ")
}

func RunCmd(cmd string, stdin io.Reader) (string, error) {
	l := strw.SplitNoEmptyKeepQuote(cmd, ' ')
	if len(l) < 1 {
		fmt.Println("cmd is empty")
		return "", errors.New("cmd is empty")
	}
	if stdin == nil {
		stdin = os.Stdin
	}
	var buf bytes.Buffer
	command := exec.Command(l[0], l[1:]...)
	command.Stdin = stdin
	command.Stderr = os.Stderr
	command.Stdout = &buf
	err := command.Run()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func RunCmdWithTimeout(cmd string, timeout time.Duration) (string, error) {
	wg := sync.WaitGroup{}
	var err error
	var res string
	wg.Add(1)
	go func(err *error) {
		res, *err = RunCmd(cmd, nil)
		defer wg.Done()
	}(&err)
	if TimeoutWait(&wg, timeout) != nil {
		return "", fmt.Errorf("timeout Execute command: %s (%v)", cmd, timeout)
	}
	return res, err
}

// Goid
func Goid() int {
	return int(goid.Get())
}

func ReadClipboardText() string {
	res, err := clipboard.ReadAll()
	if err != nil {
		panic(err)
	}
	return res
}

func WriteClipboardText(content string) {
	if err := clipboard.WriteAll(content); err != nil {
		panic(err)
	}
}
