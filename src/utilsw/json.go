package utilsw

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unsafe"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/typesw"
	"golang.org/x/exp/constraints"
)

type Json struct {
	data any

	allowComment bool
}

type JsonOption func(*Json)

func WithComment(j *Json) {
	j.allowComment = true
}

var jsonType = reflect.TypeOf(new(Json))

func NewJsonFromFile(filename string, options ...JsonOption) (*Json, error) {
	filename = ExpandUser(filename)
	b, err := os.Open(filename)
	defer b.Close()
	if err != nil {
		return nil, err
	}
	options = append(options, WithComment)
	return NewJsonFromReader(b, options...)
}

func NewJsonFromByte(data []byte) (*Json, error) {
	return NewJsonFromReader(strings.NewReader(typesw.BytesToStr(data)))
}

func NewJsonFromReader(r io.Reader, options ...JsonOption) (*Json, error) {
	options = append(options, WithComment)
	res := NewJson(nil)
	for _, option := range options {
		option(res)
	}
	var rr io.Reader
	if res.allowComment {
		f := &commentsFilter{}
		rr = FilterReader(r, f)
	} else {
		rr = r
	}

	b := make([]byte, 1)
	for {
		_, err := rr.Read(b)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if unicode.IsSpace(rune(b[0])) {
			continue
		}
		break
	}
	switch b[0] {
	case '[':
		var buf strings.Builder
		buf.WriteString(fmt.Sprintf(`{"_arr": %c`, b[0]))
		newReader := io.MultiReader(strings.NewReader(buf.String()), rr, strings.NewReader("}"))
		j, err := NewJsonFromReader(newReader)
		if err != nil {
			return nil, err
		}
		return j.GetJson("_arr"), nil
	case '{':
		m := cw.NewOrderedMapT[string, any]()
		decoder := json.NewDecoder(io.MultiReader(bytes.NewReader(b), rr))
		decoder.UseNumber()
		err := decoder.Decode(&m)
		if err != nil {
			return nil, err
		}
		res.data = m
		return res, nil
	default:
		return nil, errors.New("failed to parse json")
	}
}

func NewJsonFromString(content string, options ...JsonOption) (*Json, error) {
	return NewJsonFromReader(strings.NewReader(content), options...)
}

func NewJson(data any) *Json {
	if data == nil {
		data = cw.NewOrderedMapT[string, any]()
	}
	for isJson(data) {
		data = unwrapJson(data)
	}
	jsonArr, isJson_ := data.([]Json)
	jsonPtrArr, isJsonPtr := data.([]*Json)
	interfaceArr, isInterface := data.([]any)

	var dataArr []any
	if isJson_ {
		dataArr = unwrapArr(jsonArr)
	} else if isJsonPtr {
		dataArr = unwrapArr(jsonPtrArr)
	} else if isInterface {
		dataArr = unwrapArr(interfaceArr)
	}
	if dataArr != nil {
		data = dataArr
	}
	return &Json{data: data}
}

type type_ interface {
	bool | constraints.Float | constraints.Integer | string | Json | any
}

type keytype interface {
	string | int
}

func getByIndex[T type_](j *Json, idx int) T {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
		return *new(T)
	}
	data, ok := j.data.([]any)
	if !ok {
		fmt.Println("ERROR: json is not array")
		return *new(T)
	}
	if idx >= len(data) {
		return *new(T)
	}
	if res, ok := data[idx].(T); ok {
		return res
	}
	res := &Json{
		data: data[idx],
	}
	return *(*T)(unsafe.Pointer(&res))
}

func isJson(data any) bool {
	_, ok := data.(Json)
	if ok {
		return true
	}
	_, ok = data.(*Json)
	return ok
}

func unwrapArr[T any | Json | *Json](arr []T) []any {
	res := make([]any, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		var e any = arr[i]
		for isJson(e) {
			e = unwrapJson(e)
		}
		res = append(res, e)
	}
	return res
}

func unwrapJson(value any) any {
	data, ok := value.(Json)
	if ok {
		return data.data
	}
	ptr, ok := value.(*Json)
	if ok {
		return ptr.data
	}
	return nil
}

func getT[T type_, U keytype](j *Json, key U) T {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
		return *new(T)
	}
	data, ok := j.data.(*cw.OrderedMapT[string, any])
	if !ok {
		keyKind := reflect.TypeOf(key).Kind()
		if keyKind == reflect.Int {
			return getByIndex[T](j, int(reflect.ValueOf(key).Int()))
		}
	}
	if data == nil {
		return *new(T)
	}
	strKey := reflect.ValueOf(key).String()
	if data.GetOrDefault(strKey, nil) == nil {
		return *new(T)
	}
	// fmt.Println("??", strKey, reflect.TypeOf(data.Get(strKey)).Kind() == reflect.Pointer, reflect.TypeOf(*new(T)), jsonType)
	if res, ok := data.Get(strKey).(T); ok {
		return res
	} else if ptrRes, ok := data.Get(strKey).(*T); ok {
		return *ptrRes
	} else if k := reflect.TypeOf(data.Get(strKey)).Kind(); k == reflect.Pointer && reflect.TypeOf(*new(T)) == jsonType {
		obj, ok := data.Get(strKey).(*cw.OrderedMapT[string, any])
		if !ok {
			fmt.Printf("ERROR: key %s is not Json object, maybe array?\n", strKey)
			return *new(T)
		}
		newJson := &Json{
			data: obj,
		}
		res := *(*T)(unsafe.Pointer(&newJson))
		return res
	} else if k == reflect.Slice {
		obj, _ := data.Get(strKey).([]any)
		res := &Json{
			data: obj,
		}
		// fmt.Println("..", obj, reflect.TypeOf(obj), reflect.TypeOf(*(*T)(unsafe.Pointer(res))))
		return *(*T)(unsafe.Pointer(&res))
	} else if k == reflect.Float64 &&
		reflect.TypeOf(*new(T)).Kind() == reflect.Int &&
		math.Trunc(data.Get(strKey).(float64)) == data.Get(strKey).(float64) {
		val := int(data.Get(strKey).(float64))
		return *(*T)(unsafe.Pointer(&val))
	} else {
		// fmt.Printf("ERROR: key (\"%s\", type: \"%s\") is not type (\"%s\")\n", strKey, reflect.TypeOf(data.Get(strKey)), reflect.TypeOf(*new(T)))
		return *new(T)
	}
}

func (j *Json) Set(key string, value any) bool {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
		return false
	}
	data, ok := j.data.(*cw.OrderedMapT[string, any])
	if !ok {
		fmt.Println("ERROR: not json format, json array?")
		return false
	}
	exist := data.Contains(key)
	for isJson(value) {
		value = unwrapJson(value)
	}
	// data[key] = value
	// fmt.Println("==> put", key, value)
	data.Put(key, value)
	return exist
}

func (j *Json) Add(value any) {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
		return
	}
	arr, ok := j.data.([]any)
	if !ok {
		if m, isMap := j.data.(*cw.OrderedMapT[string, any]); !isMap || m.Size() > 0 {
			fmt.Println("ERROR: not json array.")
			return
		}
		j.data = make([]any, 0, 2)
	}
	for isJson(value) {
		value = unwrapJson(value)
	}
	j.data = append(arr, value)
}

func (j *Json) GetString(key string) string {
	return getT[string](j, key)
}

func (j *Json) GetInt(key string) int {
	return getT[int](j, key)
}

func (j *Json) GetIndex(idx int) *Json {
	res := getT[*Json](j, idx)
	return res
}

func (j *Json) GetFloat(key string) float64 {
	return getT[float64](j, key)
}

func (j *Json) GetBool(key string) bool {
	return getT[bool](j, key)
}

func (j *Json) GetJson(key string) *Json {
	if j == nil {
		return nil
	}
	if j.IsArray() {
		if keyInt, err := strconv.Atoi(key); err != nil {
			return nil
		} else {
			return j.GetIndex(keyInt)
		}
	}
	res := getT[*Json](j, key)
	if res != nil && res.data == nil {
		data := getT[*cw.OrderedMapT[string, any]](j, key)
		res.data = data
	}
	return res
}

func (j *Json) Get(key string) any {
	if j == nil {
		return nil
	}
	if j.IsArray() {
		if keyInt, err := strconv.Atoi(key); err != nil {
			return nil
		} else {
			return j.GetIndex(keyInt)
		}
	}
	return getT[any](j, key)
}

func (j *Json) GetOrDefault(key string, defaultVal any) any {
	if j == nil || j.data == nil {
		return defaultVal
	}
	if j.ContainsKey(key) {
		return j.Get(key)
	}
	return defaultVal
}

func (json *Json) flatten() []*Json {
	if !json.IsArray() {
		return []*Json{json}
	}
	var result []*Json
	for i := 0; i < json.Len(); i++ {
		sub := json.GetIndex(i)
		result = append(result, sub.flatten()...)
	}
	return result
}

func (json *Json) Extract(key string) *Json {
	if json == nil || json.data == nil {
		return nil
	}
	idx := strings.LastIndexByte(key, '.')
	if idx < 0 {
		return json.extract(key)
	}

	last := strings.TrimSpace(key[idx+1:])

	if last[0] == '[' && last[len(last)-1] == ']' {
		if !strings.Contains(last, ",") {
			return json.extract(key)
		}
		rootKey := key[:idx]
		m := cw.NewOrderedMapT[string, any]()
		for subKey := range strw.SplitByToken(strings.NewReader(last[1:len(last)-1]), ",", false) {
			sub := json.extract(fmt.Sprintf("%s.%s", rootKey, subKey))
			m.Put(subKey, sub)
		}
		var arr []*Json
		for entry := range m.Iter().Iterate() {
			subKey := entry.Key()
			absKey := fmt.Sprintf("%s.%s", rootKey, subKey)
			val := entry.Val().(*Json)
			for i, k := range val.Keys() {
				item := val.Get(k)
				if len(arr) > i {
					arr[i].Set(absKey, item)
				} else {
					subJ := NewJson(nil)
					subJ.Set(absKey, item)
					arr = append(arr, subJ)
				}
			}
		}
		return NewJson(arr)
	} else {
		return json.extract(key)
	}

}

// Extract get nested keys
func (json *Json) extract(key string) *Json {
	keys := strings.Split(key, ".")
	if len(keys) == 1 {
		if json.IsArray() {
			if json.Len() == 1 {
				return json.GetIndex(0).Extract(key)
			}
			result := NewJson(nil)
			for i := 0; i < json.Len(); i++ {
				sub := json.GetIndex(i).Extract(key)
				if sub.IsArray() {
					flatten := sub.flatten()
					for _, subJson := range flatten {
						result.Add(subJson)
					}
				} else if sub != nil && sub.data != nil {
					result.Add(sub)
				}
			}
			return result
		} else if data, ok := json.data.(*cw.OrderedMapT[string, any]); ok {
			return NewJson(data.Get(key))
		} else {
			return NewJson(nil)
		}
	}

	// result := NewJson(nil)
	currJ := json
	for _, currKey := range keys {
		currJ = currJ.Extract(currKey)
	}
	return currJ
}

func (j *Json) IsArray() bool {
	if j == nil || j.data == nil {
		return false
	}
	return reflect.TypeOf(j.data).Kind() == reflect.Slice
}

func (j *Json) Len() int {
	if j == nil || j.data == nil {
		return 0
	}
	if data, ok := j.data.(*cw.OrderedMapT[string, any]); ok {
		if data == nil {
			return 0
		}
		return data.Size()
	}
	if reflect.TypeOf(j.data).Kind() == reflect.Slice {
		return reflect.ValueOf(j.data).Len()
	}
	return 0
}

func (j *Json) Keys() []string {
	result := make([]string, 0, j.Len())
	if mResult, ok := j.data.(*cw.OrderedMapT[string, any]); ok {
		for k := range mResult.Iter().Iterate() {
			result = append(result, k.Key())
		}
		return result
	}
	if reflect.TypeOf(j.data).Kind() == reflect.Slice {
		for idx := 0; idx < reflect.ValueOf(j.data).Len(); idx++ {
			result = append(result, strconv.Itoa(idx))
		}
		return result
	}

	return nil
}

func (j *Json) ContainsKey(key string) bool {
	if j.IsArray() {
		if iKey, err := strconv.Atoi(key); err != nil {
			return false
		} else {
			return iKey >= 0 && iKey < j.Len()
		}
	}

	// json object
	if m, ok := j.data.(*cw.OrderedMapT[string, any]); ok {
		return m.Contains(key)
	}
	return false
}

func (j *Json) Scalar() any {
	if _, ok := j.data.(*cw.OrderedMapT[string, any]); ok {
		return nil
	}
	t := reflect.TypeOf(j.data)
	if t == nil || t.Kind() == reflect.Slice {
		return nil
	}
	return j.data
}

func (j *Json) String() string {
	return j.StringWithIndent("", "")
}

func (j *Json) StringWithIndent(prefix, indent string) string {
	var buf strings.Builder
	var err error
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(prefix, indent)
	err = encoder.Encode(j.data)
	if err != nil {
		return fmt.Sprintf("%v", j.data)
	}

	res := buf.String()
	// \u0026, \u003c, and \u003e
	res = strings.ReplaceAll(res, `\u0026`, "&")
	res = strings.ReplaceAll(res, `\u003c`, "<")
	res = strings.ReplaceAll(res, `\u003e`, ">")
	return res
}

func (j *Json) ToFile(fname string) {
	WriteToFile(fname, typesw.StrToBytes(j.String()))
}

func (j *Json) AbsKey(key string) []string {
	return j.absKeySearch(key, "root")
}

func (j *Json) absKeySearch(key string, currPath string) []string {
	sep := "->"
	res := cw.NewSet()
	if j.IsArray() {
		for i := 0; i < j.Len(); i++ {
			sub := j.GetIndex(i)
			paths := sub.absKeySearch(key, fmt.Sprintf("%s%s%d", currPath, sep, i))
			for _, path := range paths {
				res.Add(path)
			}
		}
	} else {
		if j.ContainsKey(key) {
			res.Add(currPath + sep + key)
		} else {
			for _, subKey := range j.Keys() {
				// fmt.Println("here", key, subKey)
				val := j.GetJson(subKey)
				if val == nil {
					continue
				}
				paths := val.absKeySearch(key, currPath+sep+subKey)
				for _, path := range paths {
					res.Add(path)
				}
			}
		}
	}
	return res.ToStringSlice()
}
