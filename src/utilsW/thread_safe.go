package utilsW

import (
	"fmt"
	"io"
	"sync"
)

var mu sync.Mutex

// Fprintf is thread safe
func Fprintf(w io.Writer, format string, a ...interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Fprintf(w, format, a...)
}

func Fprintln(w io.Writer, a ...interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Fprintln(w, a...)
}

func Fprint(w io.Writer, a ...interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Fprint(w, a...)
}

func Printf(format string, a ...interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Printf(format, a...)
}

func Println(a ...interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Println(a...)
}

func Print(a ...interface{}) (int, error) {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Print(a...)
}

func Sprintf(format string, a ...interface{}) string {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Sprintf(format, a...)
}

func Sprintln(a ...interface{}) string {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Sprintln(a...)
}

func Sprint(a ...interface{}) string {
	mu.Lock()
	defer mu.Unlock()
	return fmt.Sprint(a...)
}

type ThreadSafeVal struct {
	val interface{}
	mu  sync.RWMutex
}

func NewThreadSafeVal(val interface{}) *ThreadSafeVal {
	return &ThreadSafeVal{val, sync.RWMutex{}}
}

func (obj *ThreadSafeVal) Get() interface{} {
	obj.mu.RLock()
	defer obj.mu.RUnlock()
	return obj.val
}

func (obj *ThreadSafeVal) Set(val interface{}) {
	obj.mu.Lock()
	defer obj.mu.Unlock()
	obj.val = val
}
