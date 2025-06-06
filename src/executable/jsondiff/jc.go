package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/utilsw"
)

// var missing = utilsw.NewJson(nil)
// var extra = utilsw.NewJson(nil)
var diff = utilsw.NewJson(nil)

var diffKeys = cw.NewOrderedSet()

func buildJson(key string, old, new interface{}) *utilsw.Json {
	res := utilsw.NewJson(nil)
	res.Set("key", key)
	res.Set("old", old)
	res.Set("new", new)
	return res
}

func absKey(prefix string, key string) string {
	if key == "" {
		return prefix
	}
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func compareJson(currKey string, j1, j2 *utilsw.Json) {
	if j1 == nil && j2 == nil {
		return
	}
	if j1 == nil || j2 == nil {
		k := absKey(currKey, "")
		diff.Add(buildJson(k, j1, j2))
		diffKeys.Add(k)
		return
	}
	for _, key := range j1.Keys() {
		k := absKey(currKey, key)
		v1 := j1.Get(key)
		v2 := j2.Get(key)

		if !j2.ContainsKey(key) {
			diff.Add(buildJson(k, v1, v2))
			continue
		}

		t1 := reflect.TypeOf(v1)
		t2 := reflect.TypeOf(v2)
		if t1 != t2 {
			diff.Add(buildJson(k, v1, v2))
			continue
		}

		if _, ok := v1.(*cw.OrderedMap); ok {
			compareJson(absKey(currKey, key), utilsw.NewJson(v1), utilsw.NewJson(v2))
			continue
		}
		if _, ok := v1.([]interface{}); ok {
			compareJson(absKey(currKey, key), utilsw.NewJson(v1), utilsw.NewJson(v2))
			continue
		}
		if _, ok := v1.(*utilsw.Json); ok {
			compareJson(absKey(currKey, key), v1.(*utilsw.Json), v2.(*utilsw.Json))

			continue
		}
		// normal types
		if v1 != v2 {
			diff.Add(buildJson(k, v1, v2))
			continue
		}
	}
	if j1.Scalar() != j2.Scalar() {
		diff.Add(buildJson(currKey, j1.Scalar(), j2.Scalar()))
	}
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("jc old.json new.json")
		return
	}

	oldJson := utilsw.NewJsonFromFile(os.Args[1])
	newJson := utilsw.NewJsonFromFile(os.Args[2])
	compareJson("", oldJson, newJson)
	diff.ToFile("./_s.json")
}
