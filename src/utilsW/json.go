package utilsW

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/grewwc/go_tools/src/stringsW"
	"golang.org/x/exp/constraints"
)

type Json struct {
	data interface{}
}

func NewJsonFromFile(filename string) *Json {
	b, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	var res Json
	json.Unmarshal(b, &res.data)
	return &res
}

func NewJsonFromByte(data []byte) *Json {
	var res Json
	json.Unmarshal(data, &res.data)
	return &res
}

func NewJsonFromString(content string) *Json {
	var res Json
	json.Unmarshal(stringsW.StringToBytes(content), &res.data)
	return &res
}

func NewJson(data interface{}) *Json {
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
	if _, ok := data[strKey]; !ok {
		return *new(T)
	}
	if res, ok := data[strKey].(T); ok {
		return res
	} else if reflect.TypeOf(data[strKey]).Kind() == reflect.Map && reflect.TypeOf(*new(T)) == reflect.TypeOf(*new(Json)) {
		obj, ok := data[strKey].(map[string]interface{})
		if !ok {
			fmt.Printf("ERROR: key %s is not Json object, maybe array?\n", strKey)
			return *new(T)
		}
		newJson := &Json{
			data: obj,
		}
		return *(*T)(unsafe.Pointer(newJson))

	} else if reflect.TypeOf(data[strKey]).Kind() == reflect.Slice {
		obj, _ := data[strKey].([]interface{})
		res := Json{
			data: obj,
		}
		return *(*T)(unsafe.Pointer(&res))
	} else {
		fmt.Printf("ERROR: key (\"%s\") is not type (\"%s\")\n", strKey, reflect.TypeOf(*new(T)))
		return *new(T)
	}
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
	s, err := json.Marshal(j.data)
	if err != nil {
		return fmt.Sprintf("%v", j.data)
	}
	return stringsW.BytesToString(s)
}
