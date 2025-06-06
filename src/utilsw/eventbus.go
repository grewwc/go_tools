package utilsw

import (
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw/_utils_helpers"
)

type EventBus struct {
	m       typesw.IConcurrentMap[any, *cw.ConcurrentHashSet[string]]
	nameMap typesw.IConcurrentMap[string, *reflect.Method]
	wg      *sync.WaitGroup

	functions       *cw.ConcurrentHashSet[string]
	functionNameMap map[string]interface{}
	funcMu          *sync.RWMutex

	topics *cw.ConcurrentHashSet[string]

	parallel chan struct{}
}

func NewEventBus(n_parallel int) *EventBus {
	if n_parallel <= 0 {
		panic("n_parallel must greater than 0")
	}
	result := &EventBus{
		m:       cw.NewMutexMap[any, *cw.ConcurrentHashSet[string]](),
		nameMap: cw.NewConcurrentHashMap[string, *reflect.Method](nil, nil),
		wg:      &sync.WaitGroup{},

		functions:       cw.NewConcurrentHashSet[string](nil, nil),
		functionNameMap: make(map[string]interface{}),
		funcMu:          &sync.RWMutex{},

		topics: cw.NewConcurrentHashSet[string](nil, nil),

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
	if !b.m.Contains(listener) {
		s := cw.NewConcurrentHashSet[string](nil, nil)
		s.AddAll(methodNames...)
		b.m.Put(listener, s)
	} else {
		b.m.Get(listener).AddAll(methodNames...)
	}
	if len(methods) > 0 {
		for i := 0; i < len(methods); i++ {
			b.nameMap.Put(methodNames[i], methods[i])
		}
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

	if !b.m.Contains(listener) {
		return
	}
	methods := _utils_helpers.GetMethods(listener)
	methodNames := _utils_helpers.MethodArrToString(topic, methods)
	b.topics.Delete(topic)
	b.m.Get(listener).DeleteAll(methodNames...)

	if len(methods) > 0 {
		b.nameMap.DeleteAll(methodNames...)
	}
}

func (b *EventBus) Post(topic string, args ...interface{}) {
	// function
outer:
	for funcname := range b.functions.Iter().Iterate() {
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

	for obj := range b.m.Iter().Iterate() {
		methodNameSet := b.m.Get(obj)
	methodLoop:
		for methodName := range methodNameSet.Iter().Iterate() {
			methodName = _utils_helpers.RemoveTopicFromMethodName(topic, methodName)
			methodName = _utils_helpers.AddTopicToMethodName(topic, methodName)
			// b.nameMapMu.RLock()
			// method := b.nameMap[methodName]
			method := b.nameMap.GetOrDefault(methodName, nil)
			if method == nil {
				continue
			}
			// b.nameMapMu.RUnlock()
			in := []reflect.Value{
				reflect.ValueOf(obj),
				reflect.ValueOf(b),
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
