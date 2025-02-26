package utilsW

import (
	"reflect"
	"sync"
	"time"

	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/utilsW/_utils_helpers"
)

type Subscribe interface{}

type EventBus struct {
	m         map[interface{}]*containerW.ConcurrentSet[string]
	nameMap   map[string]*reflect.Method
	mu        *sync.RWMutex
	nameMapMu *sync.RWMutex
	wg        *sync.WaitGroup
}

func NewEventBus() *EventBus {
	return &EventBus{
		m:         make(map[interface{}]*containerW.ConcurrentSet[string]),
		nameMap:   make(map[string]*reflect.Method),
		mu:        &sync.RWMutex{},
		nameMapMu: &sync.RWMutex{},
		wg:        &sync.WaitGroup{},
	}
}

func (b *EventBus) Register(listener interface{}) {
	if listener == nil {
		return
	}
	type_ := reflect.TypeOf(listener)
	if type_ == nil {
		return
	}
	methods := _utils_helpers.GetMethods(listener)
	methodNames := _utils_helpers.MethodArrToString(methods)
	b.mu.RLock()
	if s, ok := b.m[listener]; !ok {
		b.mu.RUnlock()
		s := containerW.NewConcurrentSet(methodNames...)
		b.mu.Lock()
		b.m[listener] = s
		b.mu.Unlock()
	} else {
		b.mu.RUnlock()
		s.AddAll(methodNames...)
	}
	if len(methods) > 0 {
		b.nameMapMu.Lock()
		for i := 0; i < len(methods); i++ {
			b.nameMap[methodNames[i]] = methods[i]
		}
		b.nameMapMu.Unlock()
	}
}

func (b *EventBus) UnRegister(listener interface{}) {
	if listener == nil {
		return
	}
	type_ := reflect.TypeOf(listener)
	if type_ == nil {
		return
	}
	b.mu.RLock()
	if _, ok := b.m[listener]; !ok {
		return
	}
	methods := _utils_helpers.GetMethods(listener)
	methodNames := _utils_helpers.MethodArrToString(methods)
	b.m[listener].DeleteAll(methodNames...)
	b.mu.RUnlock()

	if len(methods) > 0 {
		b.nameMapMu.Lock()
		for _, name := range methodNames {
			delete(b.nameMap, name)
		}
		b.nameMapMu.Unlock()
	}
}

func (b *EventBus) Post(args ...interface{}) {
	b.mu.RLock()
	for obj, methodNames := range b.m {
	methodLoop:
		for methodName := range methodNames.Iterate() {
			b.nameMapMu.RLock()
			method := b.nameMap[methodName]
			b.nameMapMu.RUnlock()
			in := []reflect.Value{
				reflect.ValueOf(obj),
				reflect.ValueOf(0),
			}
			if len(args) > 0 {
				if method.Type.NumIn() <= 2 || method.Type.NumIn()-2 != len(args) {
					continue
				}
				for i := 0; i < len(args); i++ {
					if method.Type.In(i+2).String() != reflect.TypeOf(args[i]).String() {
						continue methodLoop
					}
					in = append(in, reflect.ValueOf(args[i]))
				}
			} else if method.Type.NumIn() > 2 {
				continue
			}
			b.wg.Add(1)
			go func(method *reflect.Method) {
				defer b.wg.Done()
				method.Func.Call(in)
			}(method)
		}
	}
	defer b.mu.RUnlock()
}

func (b *EventBus) Wait() {
	b.wg.Wait()
}

func (b *EventBus) WaitTimeout(duration time.Duration) error {
	return TimeoutWait(b.wg, duration)
}
