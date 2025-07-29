package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw"
)

var diff = utilsw.NewJson(nil)

var diffKeys = cw.NewOrderedSet()

func buildJson(key string, old, new interface{}) *utilsw.Json {
	res := utilsw.NewJson(nil)
	if old == nil {
		key += "  [new]"
	} else if new == nil {
		key += "  [old]"
	}
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
	keys := append(j1.Keys(), j2.Keys()...)
	for _, key := range keys {
		k := absKey(currKey, key)
		if diffKeys.Contains(k) {
			continue
		}
		v1 := j1.GetOrDefault(key, nil)
		v2 := j2.GetOrDefault(key, nil)

		if !j2.ContainsKey(key) || !j1.ContainsKey(key) {
			diff.Add(buildJson(k, v1, v2))
			diffKeys.Add(k)
			continue
		}

		t1 := reflect.TypeOf(v1)
		t2 := reflect.TypeOf(v2)
		if t1 != t2 {
			diff.Add(buildJson(k, v1, v2))
			diffKeys.Add(k)
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
			diffKeys.Add(k)
			continue
		}
	}
	if j1.Scalar() != j2.Scalar() {
		diff.Add(buildJson(currKey, j1.Scalar(), j2.Scalar()))
	}
}

func main() {
	parser := terminalw.NewParser()
	parser.String("f", "", "format json file")
	parser.ParseArgsCmd()
	positional := parser.Positional

	if parser.ContainsFlagStrict("f") {
		fname := parser.MustGetFlagVal("f")
		var text string
		if fname == "" {
			text = utilsw.ReadClipboardText()
		} else {
			text = utilsw.ReadString(fname)
		}
		formatedJ, err := utilsw.NewJsonFromString(text)
		if err != nil {
			panic(err)
		}
		formated := formatedJ.StringWithIndent("", "  ")
		if len(text) < 1024*16 {
			fmt.Println(formated)
		} else {
			fmt.Println("write file to _f.json")
			utilsw.WriteToFile(fname, typesw.StrToBytes(formated))
		}
		return
	}

	if positional.Len() != 2 {
		parser.PrintDefaults()
		fmt.Println("j old.json new.json")
		return
	}

	oldJson, err := utilsw.NewJsonFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	newJson, err := utilsw.NewJsonFromFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	compareJson("", oldJson, newJson)
	fname := "./_s.json"
	diff.ToFile(fname)
	fmt.Println("write to " + fname)
}
