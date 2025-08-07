package main

import (
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/executable/jsondiff/internal"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw"
)

var diff = utilsw.NewJson(nil)
var mu sync.Mutex

// var diffKeys = cw.NewConcurrentHashSet(typesw.CreateDefaultHash[string](), typesw.CreateDefaultCmp[string]())
var diffKeys = cw.NewMutexSet[string]()

var sort bool
var mt bool
var print bool

func addDiff(d *utilsw.Json) {
	mu.Lock()
	diff.Add(d)
	mu.Unlock()
}

func buildJson(key string, old, new any) *utilsw.Json {
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
	if mt {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		go compareJsonMtHelper(currKey, j1, j2, wg)
		wg.Wait()
	} else {
		compareJsonHelper(currKey, j1, j2)
	}
}

func compareJsonMtHelper(currKey string, j1, j2 *utilsw.Json, wg *sync.WaitGroup) {
	defer wg.Done()
	if j1 == nil || j2 == nil {
		k := absKey(currKey, "")
		addDiff(buildJson(k, j1, j2))
		diffKeys.Add(k)
		return
	}

	if j1.IsArray() && j2.IsArray() && sort {
		j1.SortArray(internal.SortJson)
		j2.SortArray(internal.SortJson)
	}

	s := cw.NewSetT(j1.Keys()...)
	s.AddAll(j2.Keys()...)
	for key := range s.Data() {
		k := absKey(currKey, key)
		if diffKeys.Contains(k) {
			continue
		}
		v1 := j1.GetOrDefault(key, nil)
		v2 := j2.GetOrDefault(key, nil)
		// fmt.Println("v1, v2", key, reflect.TypeOf(v1), reflect.TypeOf(v2))
		if !j2.ContainsKey(key) || !j1.ContainsKey(key) {
			// diff.Add(buildJson(k, v1, v2))
			addDiff(buildJson(k, v1, v2))
			diffKeys.Add(k)
			continue
		}

		t1 := reflect.TypeOf(v1)
		t2 := reflect.TypeOf(v2)
		if t1 != t2 {
			addDiff(buildJson(k, v1, v2))
			diffKeys.Add(k)
			continue
		}

		if _, ok := v1.(*cw.OrderedMapT[string, any]); ok {
			wg.Add(1)
			go compareJsonMtHelper(absKey(currKey, key), utilsw.NewJson(v1), utilsw.NewJson(v2), wg)
			continue
		}
		if _, ok := v1.([]any); ok {
			jv1, jv2 := utilsw.NewJson(v1), utilsw.NewJson(v2)
			if sort {
				jv1.SortArray(internal.SortJson)
				jv2.SortArray(internal.SortJson)
			}
			wg.Add(1)
			go compareJsonMtHelper(absKey(currKey, key), jv1, jv2, wg)
			continue
		}
		if _, ok := v1.(*utilsw.Json); ok {
			// scalar
			v1J, v2J := v1.(*utilsw.Json), v2.(*utilsw.Json)
			s1, s2 := v1J.Scalar(), v2J.Scalar()
			if (s1 != nil || s2 != nil) && (!reflect.DeepEqual(s1, s2)) {
				addDiff(buildJson(k, s1, s2))
				diffKeys.Add(k)
				continue
			}
			wg.Add(1)
			go compareJsonMtHelper(absKey(currKey, key), v1J, v2J, wg)
			continue
		}
		// normal types
		if v1 != v2 {
			addDiff(buildJson(k, v1, v2))
			diffKeys.Add(k)
			continue
		}
	}
	s1, s2 := j1.Scalar(), j2.Scalar()
	if !reflect.DeepEqual(s1, s2) {
		addDiff(buildJson(currKey, s1, s2))
	}
}

func compareJsonHelper(currKey string, j1, j2 *utilsw.Json) {
	if j1 == nil || j2 == nil {
		k := absKey(currKey, "")
		diff.Add(buildJson(k, j1, j2))
		diffKeys.Add(k)
		return
	}

	if j1.IsArray() && j2.IsArray() && sort {
		j1.SortArray(internal.SortJson)
		j2.SortArray(internal.SortJson)
	}

	s := cw.NewOrderedSetT(j1.Keys()...)
	s.AddAll(j2.Keys()...)
	for _, key := range s.Data().Keys() {
		k := absKey(currKey, key)
		// fmt.Println("=====>", k)
		if diffKeys.Contains(k) {
			continue
		}
		v1 := j1.GetOrDefault(key, nil)
		v2 := j2.GetOrDefault(key, nil)
		// fmt.Println("v1, v2", key, reflect.TypeOf(v1), reflect.TypeOf(v2))
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

		if _, ok := v1.(*cw.OrderedMapT[string, any]); ok {
			compareJsonHelper(absKey(currKey, key), utilsw.NewJson(v1), utilsw.NewJson(v2))
			continue
		}
		if _, ok := v1.([]any); ok {
			jv1, jv2 := utilsw.NewJson(v1), utilsw.NewJson(v2)
			if sort {
				jv1.SortArray(internal.SortJson)
				jv2.SortArray(internal.SortJson)
			}
			compareJsonHelper(absKey(currKey, key), jv1, jv2)
			continue
		}
		if _, ok := v1.(*utilsw.Json); ok {
			// scalar
			v1J, v2J := v1.(*utilsw.Json), v2.(*utilsw.Json)
			s1, s2 := v1J.Scalar(), v2J.Scalar()
			if (s1 != nil || s2 != nil) && (!reflect.DeepEqual(s1, s2)) {
				diff.Add(buildJson(k, s1, s2))
				diffKeys.Add(k)
				continue
			}

			compareJsonHelper(absKey(currKey, key), v1J, v2J)
			continue
		}
		// normal types
		if v1 != v2 {
			diff.Add(buildJson(k, v1, v2))
			diffKeys.Add(k)
			continue
		}
	}
	s1, s2 := j1.Scalar(), j2.Scalar()
	if !reflect.DeepEqual(s1, s2) {
		diff.Add(buildJson(currKey, s1, s2))
	}
}

func main() {
	parser := terminalw.NewParser()
	parser.String("f", "", "format json file")
	parser.Bool("sort", false, "sort slice, ignore slice order")
	parser.Bool("mt", false, "multi-thread, result is not ordered.")
	parser.Bool("p", false, "print to console.")
	parser.ParseArgsCmd("sort", "mt", "p")
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

	sort = parser.ContainsFlagStrict("sort")
	mt = parser.ContainsFlagStrict("mt")
	print = parser.ContainsFlagStrict("p")
	oldJson, err := utilsw.NewJsonFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	newJson, err := utilsw.NewJsonFromFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	compareJson("", oldJson, newJson)
	if print {
		fmt.Println(diff.StringWithIndent("", "  "))
	}
	fname := "./_s.json"
	diff.ToFile(fname)
	fmt.Println("write to " + fname)
}
