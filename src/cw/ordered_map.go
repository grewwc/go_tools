package cw

import (
	"container/list"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/grewwc/go_tools/src/typesw"
)

// OrderedMap is a map that maintains the order of insertion.
type OrderedMap struct {
	m map[interface{}]*list.Element
	l *list.List
}

const (
	cap = 1
)

type MapEntry[K, V any] struct {
	k K
	v V
}

func (entry *MapEntry[K, V]) Key() K {
	return entry.k
}

func (entry *MapEntry[K, V]) Val() V {
	return entry.v
}

func NewOrderedMap() *OrderedMap {
	l := list.New()
	res := &OrderedMap{make(map[interface{}]*list.Element, cap), l}
	return res
}

func (s *OrderedMap) Put(k, v interface{}) {
	if node, exist := s.m[k]; !exist {
		e := s.l.PushBack(&MapEntry[any, any]{k, v})
		s.m[k] = e
	} else {
		val := node.Value.(*MapEntry[any, any])
		val.k = k
		val.v = v
	}
}

func (s *OrderedMap) PutIfAbsent(k, v interface{}) {
	if _, ok := s.m[k]; ok {
		return
	}
	s.Put(k, v)
}

func (s *OrderedMap) Get(k interface{}) interface{} {
	if s.m[k] == nil {
		return nil
	}
	return s.m[k].Value.(*MapEntry[any, any]).v
}

func (s *OrderedMap) GetOrDefault(k, defaultVal interface{}) interface{} {
	if val, ok := s.m[k]; ok {
		return val.Value.(*MapEntry[any, any]).v
	}
	return defaultVal
}

func (s OrderedMap) Iter() typesw.IterableT[*MapEntry[any, any]] {
	return &listIterator[*MapEntry[any, any]]{
		data:    s.l,
		reverse: false,
	}
}

func (s *OrderedMap) Contains(k interface{}) bool {
	if _, exist := s.m[k]; exist {
		return true
	}
	return false
}

func (s *OrderedMap) Delete(k interface{}) bool {
	if val, ok := s.m[k]; ok {
		delete(s.m, k)
		s.l.Remove(val)
		return true
	}
	return false
}

func (s *OrderedMap) DeleteAll(ks ...interface{}) {
	for _, k := range ks {
		s.Delete(k)
	}
}

func (s *OrderedMap) Empty() bool {
	return len(s.m) == 0
}

func (s *OrderedMap) Size() int {
	return len(s.m)
}

func (s *OrderedMap) Clear() {
	s.m = make(map[interface{}]*list.Element, cap)
	s.l.Init()
}

func (om *OrderedMap) parseobject(dec *json.Decoder) (err error) {
	var t json.Token
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		key, ok := t.(string)
		if !ok {
			return fmt.Errorf("expecting JSON key should be always a string: %T: %v", t, t)
		}

		t, err = dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return err
		}

		// om.keys = append(om.keys, key)
		// om.keys[key] = om.l.PushBack(key)
		// om.m[key] = value
		om.Put(key, value)
	}

	t, err = dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '}' {
		return fmt.Errorf("expect JSON object close with '}'")
	}

	return nil
}

func parsearray(dec *json.Decoder) (arr []interface{}, err error) {
	var t json.Token
	arr = make([]interface{}, 0, cap)
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return
		}

		var value interface{}
		value, err = handledelim(t, dec)
		if err != nil {
			return
		}
		arr = append(arr, value)
	}
	t, err = dec.Token()
	if err != nil {
		return
	}
	if delim, ok := t.(json.Delim); !ok || delim != ']' {
		err = fmt.Errorf("expect JSON array close with ']'")
		return
	}

	return
}

func handledelim(t json.Token, dec *json.Decoder) (res interface{}, err error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			om2 := NewOrderedMap()
			err = om2.parseobject(dec)
			if err != nil {
				return
			}
			return om2, nil
		case '[':
			var value []interface{}
			value, err = parsearray(dec)
			if err != nil {
				return
			}
			return value, nil
		default:
			return nil, fmt.Errorf("unexpected delimiter: %q", delim)
		}
	}
	return t, nil
}

var pool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

func encode(w io.Writer, om *OrderedMap) error {
	buf := pool.Get().(*strings.Builder)
	buf.Reset()
	defer pool.Put(buf)
	buf.WriteString("{")
	i := 0
	for e := om.l.Front(); e != nil; e = e.Next() {
		item := e.Value.(*MapEntry[any, any])
		// 写入键名
		buf.WriteByte('"')
		buf.WriteString(item.Key().(string))
		buf.WriteByte('"')
		buf.WriteByte(':')
		// fmt.Println(",,", item.Key(), item.Val(), reflect.TypeOf(item.Val()))
		switch v := item.Val().(type) {
		case *OrderedMap:
			subBuf := pool.Get().(*strings.Builder)
			subBuf.Reset()
			if err := encode(subBuf, v); err != nil {
				return err
			}
			buf.WriteString(subBuf.String())
			pool.Put(subBuf)
		case []interface{}:
			buf.WriteByte('[')
			for j, elem := range v {
				if omElem, ok := elem.(*OrderedMap); ok {
					// subBuf := &strings.Builder{}
					subBuf := pool.Get().(*strings.Builder)
					subBuf.Reset()
					if err := encode(subBuf, omElem); err != nil {
						pool.Put(subBuf)
						return err
					}
					buf.WriteString(subBuf.String())
					pool.Put(subBuf)
				} else {
					encoded, err := json.Marshal(elem)
					if err != nil {
						return err
					}
					buf.Write(encoded)
				}
				if j < len(v)-1 {
					buf.WriteByte(',')
				}
			}
			buf.WriteByte(']')
		default:
			encoded, err := json.Marshal(v)
			if err != nil {
				return err
			}
			buf.Write(encoded)
		}
		// 添加逗号分隔符
		if i < om.Size()-1 {
			buf.WriteByte(',')
		}
		i++
	}

	buf.WriteByte('}')
	_, err := w.Write(typesw.StrToBytes(buf.String()))
	return err
}

func (om *OrderedMap) MarshalJSON() (res []byte, err error) {
	var buf strings.Builder
	buf.Grow(len(res))
	if err := encode(&buf, om); err != nil {
		return nil, err
	}
	return typesw.StrToBytes(buf.String()), nil
}

// this implements type json.Unmarshaler interface, so can be called in json.Unmarshal(data, om)
func (om *OrderedMap) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(strings.NewReader(typesw.BytesToStr(data)))
	dec.UseNumber()

	// must open with a delim token '{'
	t, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := t.(json.Delim); !ok || delim != '{' {
		return fmt.Errorf("expect JSON object open with '{'")
	}

	err = om.parseobject(dec)
	if err != nil {
		return err
	}

	t, err = dec.Token()
	if err != io.EOF {
		return fmt.Errorf("expect end of JSON object but got more token: %T: %v or err: %v", t, t, err)
	}

	return nil
}

func (s *OrderedMap) String() string {
	if s == nil || s.m == nil || s.l == nil {
		return ""
	}
	res := make([]interface{}, 0, len(s.m))
	front := s.l.Front()
	if front == nil {
		return ""
	}
	for i := 0; i < s.Size(); i++ {
		k := front.Value.(*MapEntry[any, any]).k
		res = append(res, fmt.Sprintf("%v: %v", k, s.Get(k)))
		front = front.Next()
	}

	return fmt.Sprintf("%v", res)
}

func (s OrderedMap) ShallowCopy() *OrderedMap {
	result := NewOrderedMap()
	front := s.l.Front()
	if front == nil {
		return result
	}
	for i := 0; i < s.Size(); i++ {
		k := front.Value.(*MapEntry[any, any]).k
		result.Put(k, s.Get(k))
		front = front.Next()
	}
	return result
}
