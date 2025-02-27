package utilsW

import (
	"reflect"
	"runtime"
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

	functions       *containerW.ConcurrentSet[string]
	functionNameMap map[string]interface{}
	funcMu          *sync.RWMutex

	topics *containerW.ConcurrentSet[string]

	parallel chan struct{}
}

func NewEventBus(n_parallel int) *EventBus {
	if n_parallel <= 0 {
		panic("n_parallel must greater than 0")
	}
	result := &EventBus{
		m:         make(map[interface{}]*containerW.ConcurrentSet[string]),
		nameMap:   make(map[string]*reflect.Method),
		mu:        &sync.RWMutex{},
		nameMapMu: &sync.RWMutex{},
		wg:        &sync.WaitGroup{},

		functions:       containerW.NewConcurrentSet[string](),
		functionNameMap: make(map[string]interface{}),
		funcMu:          &sync.RWMutex{},

		topics: containerW.NewConcurrentSet[string](),

		parallel: make(chan struct{}, n_parallel),
	}
	runtime.SetFinalizer(result, func(obj *EventBus) {
		if obj != nil {
			close(obj.parallel)
		}
	})
	return result
}

func (b *EventBus) Register(topic string, listener interface{}) {
	if listener == nil {
		return
	}
	type_ := reflect.TypeOf(listener)
	if type_ == nil {
		return
	}
	// function
	if type_.Kind() == reflect.Func {
		funcName := _utils_helpers.GetFunctionName(listener)
		funcName = _utils_helpers.AddTopicToMethodName(topic, funcName)
		b.functions.Add(funcName)
		b.topics.Add(topic)
		b.funcMu.Lock()
		b.functionNameMap[funcName] = listener
		b.funcMu.Unlock()
		return
	}
	methods := _utils_helpers.GetMethods(listener)
	methodNames := _utils_helpers.MethodArrToString(topic, methods)
	b.topics.Add(topic)
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

func (b *EventBus) UnRegister(topic string, listener interface{}) {
	if listener == nil {
		return
	}
	type_ := reflect.TypeOf(listener)
	if type_ == nil {
		return
	}
	// function
	if type_.Kind() == reflect.Func {
		funcName := _utils_helpers.AddTopicToMethodName(topic, _utils_helpers.GetFunctionName(listener))
		b.functions.Delete(funcName)
		b.funcMu.Lock()
		delete(b.functionNameMap, funcName)
		b.funcMu.Unlock()
		return
	}

	b.mu.RLock()
	if _, ok := b.m[listener]; !ok {
		return
	}
	methods := _utils_helpers.GetMethods(listener)
	methodNames := _utils_helpers.MethodArrToString(topic, methods)
	b.topics.Delete(topic)
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

func (b *EventBus) Post(topic string, args ...interface{}) {
	// function
outer:
	for funcname := range b.functions.Iterate() {
		funcname = _utils_helpers.RemoveTopicFromMethodName(topic, funcname)
		funcname = _utils_helpers.AddTopicToMethodName(topic, funcname)
		b.funcMu.RLock()
		ifunc := b.functionNameMap[funcname]
		tfunc := reflect.TypeOf(ifunc)
		if ifunc == nil || len(args) != tfunc.NumIn() {
			continue
		}
		// check arg types
		for i := 0; i < len(args); i++ {
			if reflect.TypeOf(args[i]).String() != tfunc.In(i).String() {
				continue outer
			}
		}
		b.wg.Add(1)
		go func() {
			b.parallel <- struct{}{}
			defer func() {
				<-b.parallel
				b.wg.Done()
			}()
			reflect.ValueOf(ifunc).Call(_utils_helpers.InterfaceToValue(args...))
		}()
		b.funcMu.RUnlock()
		return
	}

	b.mu.RLock()
	for obj, methodNameSet := range b.m {
	methodLoop:
		for methodName := range methodNameSet.Iterate() {
			methodName = _utils_helpers.RemoveTopicFromMethodName(topic, methodName)
			methodName = _utils_helpers.AddTopicToMethodName(topic, methodName)
			b.nameMapMu.RLock()
			method := b.nameMap[methodName]
			if method == nil {
				continue
			}
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
				b.parallel <- struct{}{}
				defer func() {
					<-b.parallel
					b.wg.Done()
				}()
				method.Func.Call(in)
			}(method)
		}
	}
	defer b.mu.RUnlock()
}

func (b *EventBus) BroadCast(args ...interface{}) {
	for _, topic := range b.ListTopics() {
		b.Post(topic, args...)
	}
}

func (b *EventBus) Wait() {
	b.wg.Wait()
}

func (b *EventBus) WaitTimeout(duration time.Duration) error {
	return TimeoutWait(b.wg, duration)
}

func (b *EventBus) ListTopics() []string {
	return b.topics.ToSlice()
}
