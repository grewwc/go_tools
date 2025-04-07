package utilsW

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/grewwc/go_tools/src/strW"
	"github.com/grewwc/go_tools/src/typesW"
	"golang.org/x/exp/constraints"
)

type Json struct {
	data interface{}
}

func NewJsonFromFile(filename string) *Json {
	filename = ExpandUser(filename)
	b, err := os.Open(filename)
	if err != nil {
		log.Fatalln(err)
	}
	var res Json
	d := json.NewDecoder(b)
	if err = d.Decode(&res.data); err != nil {
		panic(err)
	}
	return &res
}

func NewJsonFromByte(data []byte) *Json {
	var res Json
	if err := json.Unmarshal(data, &res.data); err != nil {
		panic(strW.SubStringQuiet(typesW.BytesToStr(data), 0, 128))
	}
	return &res
}

func NewJsonFromString(content string) *Json {
	var res Json
	if err := json.Unmarshal(typesW.StrToBytes(content), &res.data); err != nil {
		panic(err)
	}
	return &res
}

func NewJson(data interface{}) *Json {
	if data == nil {
		data = make(map[string]interface{})
	}
	for isJson(data) {
		data = unwrapJson(data)
	}
	jsonArr, isJson_ := data.([]Json)
	jsonPtrArr, isJsonPtr := data.([]*Json)
	interfaceArr, isInterface := data.([]interface{})

	var dataArr []interface{}
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
	bool | constraints.Float | constraints.Integer | string | Json | interface{}
}

type keytype interface {
	string | int
}

func getByIndex[T type_](j *Json, idx int) T {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
		return *new(T)
	}
	data, ok := j.data.([]interface{})
	if !ok {
		fmt.Println("ERROR: json is not array")
		return *new(T)
	}
	if idx >= len(data) {
		fmt.Printf("ERROR: idx (%d) >= size (%d)\n", idx, len(data))
		return *new(T)
	}
	if res, ok := data[idx].(T); ok {
		return res
	}
	res := Json{
		data: data[idx],
	}
	return *(*T)(unsafe.Pointer(&res))
}

func isJson(data interface{}) bool {
	_, ok := data.(Json)
	if ok {
		return true
	}
	_, ok = data.(*Json)
	return ok
}

func unwrapArr[T interface{} | Json | *Json](arr []T) []interface{} {
	res := make([]interface{}, 0, len(arr))
	for i := 0; i < len(arr); i++ {
		var e interface{} = arr[i]
		for isJson(e) {
			e = unwrapJson(e)
		}
		res = append(res, e)
	}
	return res
}

func unwrapJson(value interface{}) interface{} {
	data, ok := value.(Json)
	if ok {
		return data.data
	}
	ptr, jsonPtr := value.(*Json)
	if jsonPtr {
		return ptr.data
	}
	return nil
}

func getT[T type_, U keytype](j *Json, key U) T {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
		return *new(T)
	}
	data, ok := j.data.(map[string]interface{})
	if !ok {
		keyKind := reflect.TypeOf(key).Kind()
		if keyKind == reflect.Int {
			return getByIndex[T](j, int(reflect.ValueOf(key).Int()))
		}
	}
	strKey := reflect.ValueOf(key).String()
	if val, ok := data[strKey]; !ok || val == nil {
		return *new(T)
	}
	if res, ok := data[strKey].(T); ok {
		return res
	} else if ptrRes, ok := data[strKey].(*T); ok {
		return *ptrRes
	} else if k := reflect.TypeOf(data[strKey]).Kind(); k == reflect.Map && reflect.TypeOf(*new(T)) == reflect.TypeOf(*new(Json)) {
		obj, ok := data[strKey].(map[string]interface{})
		if !ok {
			fmt.Printf("ERROR: key %s is not Json object, maybe array?\n", strKey)
			return *new(T)
		}
		newJson := &Json{
			data: obj,
		}
		return *(*T)(unsafe.Pointer(newJson))

	} else if k == reflect.Slice {
		obj, _ := data[strKey].([]interface{})
		res := Json{
			data: obj,
		}
		return *(*T)(unsafe.Pointer(&res))
	} else if k == reflect.Float64 &&
		reflect.TypeOf(*new(T)).Kind() == reflect.Int &&
		math.Trunc(data[strKey].(float64)) == data[strKey].(float64) {
		val := int(data[strKey].(float64))
		return *(*T)(unsafe.Pointer(&val))
	} else {
		fmt.Printf("ERROR: key (\"%s\", type: \"%s\") is not type (\"%s\")\n", strKey, reflect.TypeOf(data[strKey]), reflect.TypeOf(*new(T)))
		return *new(T)
	}
}

func (j *Json) Set(key string, value interface{}) bool {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
		return false
	}
	data, ok := j.data.(map[string]interface{})
	if !ok {
		fmt.Println("ERROR: not json format, json array?")
		return false
	}
	_, exist := data[key]
	for isJson(value) {
		value = unwrapJson(value)
	}
	data[key] = value
	return exist
}

func (j *Json) Add(value interface{}) {
	if j == nil || j.data == nil {
		fmt.Println("ERROR: json is nil")
	}
	arr, ok := j.data.([]interface{})
	if !ok {
		if m, isMap := j.data.(map[string]interface{}); !isMap || len(m) > 0 {
			fmt.Println("ERROR: not json array.")
			return
		}
		j.data = make([]interface{}, 0, 2)
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
	res := getT[Json](j, idx)
	return &res
}

func (j *Json) GetFloat(key string) float64 {
	return getT[float64](j, key)
}

func (j *Json) GetBool(key string) bool {
	return getT[bool](j, key)
}

func (j *Json) GetJson(key string) *Json {
	res := getT[Json](j, key)
	if res.data == nil {
		return nil
	}
	return &res
}

func (j *Json) Len() int {
	if j == nil || j.data == nil {
		return 0
	}
	if data, ok := j.data.(map[string]interface{}); ok {
		return len(data)
	}
	if data, ok := j.data.([]interface{}); ok {
		return len(data)
	}
	return 0
}

func (j *Json) Keys() []string {
	result := make([]string, 0, j.Len())
	if mResult, ok := j.data.(map[string]interface{}); ok {
		for k := range mResult {
			result = append(result, k)
		}
		return result
	}
	if sliceResult, ok := j.data.([]interface{}); ok {
		for idx := range sliceResult {
			result = append(result, strconv.Itoa(idx))
		}
		return result
	}

	return nil
}

func (j *Json) ContainsKey(key string) bool {
	keys := j.Keys()
	for _, k := range keys {
		if key == k {
			return true
		}
	}
	return false
}

func (j Json) String() string {
	return j.StringWithIndent("", "")
}

func (j Json) StringWithIndent(prefix, indent string) string {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(prefix, indent)
	err := encoder.Encode(j.data)
	if err != nil {
		return fmt.Sprintf("%v", j.data)
	}
	return buf.String()
}
