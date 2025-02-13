package utilsW

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/atotto/clipboard"
	"github.com/chzyer/readline"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/stringsW"
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
	s := containerW.NewSet()
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
	size := stringsW.SplitNoEmpty(strings.TrimSpace(string(out)), " ")
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
	// 创建 readline 实例
	rl, err := readline.NewEx(&readline.Config{
		EOFPrompt:       "^D",
		Prompt:          msg,
		Stdin:           os.Stdin,
		Stderr:          os.Stderr,
		InterruptPrompt: "^C",
	})
	if err != nil {
		panic(err)
	}
	rl.Operation.ExitCompleteMode(true)
	// defer rl.Close()
	var lines []string
	for {
		line, err := rl.Readline()
		if err != nil && err == readline.ErrInterrupt {
			fmt.Println("Exit.")
			os.Exit(0)
		}
		if !multiline {
			return line
		}
		if err != nil && err == io.EOF {
			break
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func kill(cmd *exec.Cmd) error {
	kill := exec.Command("TASKKILL", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	kill.Stderr = os.Stderr
	kill.Stdout = os.Stdout
	return kill.Run()
}

// GetCommandList seperate cmd to []string
func GetCommandList(cmd string) []string {
	cmd = strings.ReplaceAll(cmd, ",", " ")
	return stringsW.SplitNoEmpty(cmd, " ")
}

func RunCmd(cmd string, stdin io.Reader) (string, error) {
	l := stringsW.SplitNoEmptyKeepQuote(cmd, ' ')
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

// copied from stackoverflow: https://stackoverflow.com/questions/75361134/how-can-i-get-a-goroutines-runtime-id
func Goid() int {
	buf := make([]byte, 32)
	n := runtime.Stack(buf, false)
	buf = buf[:n]
	// goroutine 1 [running]: ...

	buf, ok := bytes.CutPrefix(buf, stringsW.StringToBytes("goroutine "))
	if !ok {
		return -1
	}

	i := bytes.IndexByte(buf, ' ')
	if i < 0 {
		return -1
	}

	res, err := strconv.Atoi(string(buf[:i]))
	if err != nil {
		return -1
	}
	return res
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
