package configW

import (
	"io/ioutil"
	"log"
	"strings"
)

func replaceVar(str *string) (string, bool) {
	match := reVar.FindStringSubmatch(*str)
	if match != nil {
		if _, exist := variablesMap[match[2]]; !exist {
			log.Fatalf("variable %q is not set\n", match[2])
		}
		old := *str
		*str = strings.ReplaceAll(*str, match[1], variablesMap[match[2]])
		return old, true
	}

	return *str, false
}

func lsDir(fname string) []string {
	infos, err := ioutil.ReadDir(fname)
	if err != nil {
		log.Fatal(err)
	}
	res := make([]string, len(infos))
	for i, info := range infos {
		res[i] = info.Name()
	}
	return res
}

func trimSpace(strs []string) {
	for i := range strs {
		strs[i] = strings.TrimSpace(strs[i])
	}
}
