package utilsW

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"

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
			fmt.Println("ere", val)
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
	var size []string
	size = stringsW.SplitNoEmpty(strings.TrimSpace(string(out)), " ")
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

func RunCmd(cmd string) error {
	l := stringsW.SplitNoEmpty(cmd, " ")
	if len(l) < 1 {
		fmt.Println("cmd is empty")
		return errors.New("cmd is empty")
	}
	command := exec.Command(l[0], l[1:]...)
	command.Stdin = os.Stdin
	command.Stderr = os.Stderr
	stdout, _ := command.StdoutPipe()
	err := command.Start()
	if err != nil {
		return err
	}
	// return stdout.String(), nil
	for {
		buf := make([]byte, 1024)
		_, err := stdout.Read(buf)
		if err != nil {
			break
		}
		fmt.Print(string(buf))
	}
	return nil
}

func RunCmdWithTimeout(cmd string, timeout time.Duration) error {
	wg := sync.WaitGroup{}
	var err error
	wg.Add(1)
	go func(err *error) {
		*err = RunCmd(cmd)
		defer wg.Done()
	}(&err)
	if TimeoutWait(&wg, timeout) != nil {
		return fmt.Errorf("timeout Execute command: %s (%v)", cmd, timeout)
	}
	return err
}
