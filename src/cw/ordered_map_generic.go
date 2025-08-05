package cw

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/grewwc/go_tools/src/typesw"
)

// OrderedMapT is a map that maintains the order of insertion.
type OrderedMapT[K comparable, V any] struct {
	m map[K]*ListNode[*MapEntry[K, V]]
	l *LinkedList[*MapEntry[K, V]]
}

func NewOrderedMapT[K comparable, V any]() *OrderedMapT[K, V] {
	l := NewLinkedList[*MapEntry[K, V]]()
	res := &OrderedMapT[K, V]{
		l: l,
		m: make(map[K]*ListNode[*MapEntry[K, V]]),
	}
	return res
}

func (s *OrderedMapT[Key, Val]) Put(k Key, v Val) {
	if node, exist := s.m[k]; !exist {
		e := s.l.PushBack(&MapEntry[Key, Val]{k, v})
		s.m[k] = e
	} else {
		val := node.Value()
		val.k = k
		val.v = v
	}
}

func (s *OrderedMapT[Key, Val]) PutIfAbsent(k Key, v Val) {
	if _, ok := s.m[k]; ok {
		return
	}
	s.Put(k, v)
}

func (s *OrderedMapT[Key, Val]) Get(k Key) Val {
	if s.m[k] == nil {
		return *new(Val)
	}
	return s.m[k].Value().v
}

func (s *OrderedMapT[Key, Val]) GetOrDefault(k Key, defaultVal Val) Val {
	if val, ok := s.m[k]; ok {
		return val.Value().v
	}
	return defaultVal
}

func (s *OrderedMapT[Key, Val]) Keys() []Key {
	res := make([]Key, 0, len(s.m))
	for key := range s.m {
		res = append(res, key)
	}
	return res
}

func (s *OrderedMapT[Key, Val]) Iter() typesw.IterableT[*MapEntry[Key, Val]] {
	return typesw.FuncToIterable(func() chan *MapEntry[Key, Val] {
		ch := make(chan *MapEntry[Key, Val])
		go func() {
			defer close(ch)
			for curr := s.l.Front(); curr != nil; curr = curr.Next() {
				ch <- curr.Value()
			}
		}()
		return ch
	})
}

func (s *OrderedMapT[K, V]) ForEach(f func(e *MapEntry[K, V])) {
	for curr := s.l.Front(); curr != nil; curr = curr.Next() {
		f(curr.Value())
	}
}

func (s *OrderedMapT[Key, Val]) Contains(k Key) bool {
	if _, exist := s.m[k]; exist {
		return true
	}
	return false
}

func (s *OrderedMapT[Key, Val]) Delete(k Key) bool {
	if val, ok := s.m[k]; ok {
		delete(s.m, k)
		s.l.Remove(val)
		return true
	}
	return false
}

func (s *OrderedMapT[Key, Val]) DeleteAll(ks ...Key) {
	for _, k := range ks {
		s.Delete(k)
	}
}

func (s *OrderedMapT[Key, Val]) Empty() bool {
	return len(s.m) == 0
}

func (s *OrderedMapT[Key, Val]) Size() int {
	return len(s.m)
}

func (s *OrderedMapT[Key, Val]) Clear() {
	s.m = make(map[Key]*ListNode[*MapEntry[Key, Val]], cap)
	s.l.Clear()
}

func (om *OrderedMapT[Key, Val]) parseobject(dec *json.Decoder) (err error) {
	var t json.Token
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return err
		}

		key, ok := t.(Key)
		if !ok {
			return fmt.Errorf("expecting JSON key should be always a string: %T: %v", t, t)
		}

		t, err = dec.Token()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		var value any
		value, err = handledelimV2[Key, Val](t, dec)
		if err != nil {
			return err
		}

		// om.keys = append(om.keys, key)
		// om.keys[key] = om.l.PushBack(key)
		// om.m[key] = value
		if value != nil {
			om.Put(key, value.(Val))
		} else {
			om.Put(key, *new(Val))
		}
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

func parsearrayV2[Key comparable, Val any](dec *json.Decoder) (arr []any, err error) {
	var t json.Token
	arr = make([]any, 0, cap)
	for dec.More() {
		t, err = dec.Token()
		if err != nil {
			return
		}

		var value any
		value, err = handledelimV2[Key, Val](t, dec)
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

func handledelimV2[Key comparable, Val any](t json.Token, dec *json.Decoder) (res any, err error) {
	if delim, ok := t.(json.Delim); ok {
		switch delim {
		case '{':
			om2 := NewOrderedMapT[Key, Val]()
			err = om2.parseobject(dec)
			if err != nil {
				return
			}
			return om2, nil
		case '[':
			var value []any
			value, err = parsearrayV2[Key, Val](dec)
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

func encodeV2[Key comparable, Val any](w io.Writer, om *OrderedMapT[Key, Val]) error {
	buf := pool.Get().(*strings.Builder)
	buf.Reset()
	defer pool.Put(buf)
	buf.WriteString("{")
	i := 0
	for e := om.l.Front(); e != nil; e = e.Next() {
		item := e.Value()
		// 写入键名
		buf.WriteByte('"')
		fmt.Fprintf(buf, "%v", item.Key())
		buf.WriteByte('"')
		buf.WriteByte(':')
		var iItem any = item.Val()
		// fmt.Println(",,", item.Key(), item.Val(), reflect.TypeOf(item.Val()))
		switch v := iItem.(type) {
		case *OrderedMapT[Key, Val]:
			subBuf := pool.Get().(*strings.Builder)
			subBuf.Reset()
			if err := encodeV2(subBuf, v); err != nil {
				return err
			}
			buf.WriteString(subBuf.String())
			pool.Put(subBuf)
		case []any:
			buf.WriteByte('[')
			for j, elem := range v {
				if omElem, ok := elem.(*OrderedMapT[Key, Val]); ok {
					// subBuf := &strings.Builder{}
					subBuf := pool.Get().(*strings.Builder)
					subBuf.Reset()
					if err := encodeV2(subBuf, omElem); err != nil {
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

func (om *OrderedMapT[Key, Val]) MarshalJSON() (res []byte, err error) {
	var buf strings.Builder
	buf.Grow(len(res))
	if err := encodeV2(&buf, om); err != nil {
		return nil, err
	}
	return typesw.StrToBytes(buf.String()), nil
}

// this implements type json.Unmarshaler interface, so can be called in json.Unmarshal(data, om)
func (om *OrderedMapT[Key, Val]) UnmarshalJSON(data []byte) error {
	dec := json.NewDecoder(bufio.NewReader(strings.NewReader(typesw.BytesToStr(data))))
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

func (s *OrderedMapT[Key, Val]) String() string {
	if s == nil || s.m == nil || s.l == nil {
		return ""
	}
	res := make([]any, 0, len(s.m))
	front := s.l.Front()
	if front == nil {
		return ""
	}
	for i := 0; i < s.Size(); i++ {
		k := front.Value().k
		res = append(res, fmt.Sprintf("%v: %v", k, s.Get(k)))
		front = front.Next()
	}

	return fmt.Sprintf("%v", res)
}

func (s OrderedMapT[Key, Val]) ShallowCopy() *OrderedMapT[Key, Val] {
	result := NewOrderedMapT[Key, Val]()
	front := s.l.Front()
	if front == nil {
		return result
	}
	for i := 0; i < s.Size(); i++ {
		k := front.Value().k
		result.Put(k, s.Get(k))
		front = front.Next()
	}
	return result
}
