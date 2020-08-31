package configW

import (
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

func trimSpace(strs []string) {
	for i := range strs {
		strs[i] = strings.TrimSpace(strs[i])
	}
}
