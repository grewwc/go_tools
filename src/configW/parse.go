package configW

import (
	"bufio"
	"fmt"
	"go_tools/src/containerW"
	"go_tools/src/stringsW"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var supportedAttrs = containerW.NewSet()
var parsedAttrValueEachKey = containerW.NewSet()
var paresdAttrKey = containerW.NewSet()

var variablesMap map[string]string
var mapping map[string]string
var parsedAttrsMap map[string][]string

func readAttrFromFile(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		log.Printf(`no "*.%s" file\n`, attrFile)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		supportedAttrs.Add(line)
	}
}

func initSupportedAttrs(syncdir string) {
	supportedAttrs.Add("ignore")
	allFiles := lsDir(syncdir)
	var allAttrFiles []string
	for _, file := range allFiles {
		if filepath.Ext(file) == attrFile {
			allAttrFiles = append(allAttrFiles, file)
		}
	}

	for _, filename := range allAttrFiles {
		readAttrFromFile(filepath.Join(syncdir, filename))
	}

}

func initialize(syncdir string) {
	homedir := os.Getenv("HOME")
	if homedir == "" {
		log.Fatal("home dir is empty")
	}

	initSupportedAttrs(syncdir)

	// init variables
	variablesMap = make(map[string]string)

	// init parsedAttrs
	parsedAttrsMap = make(map[string][]string)

	// init mapping
	mapping = extractAll(syncdir)

	// replace variables
	for k, v := range variablesMap {
		for {
			if _, replaced := replaceVar(&v); replaced {
				if reVar.MatchString(v) {
					continue
				} else {
					variablesMap[k] = v
				}
			} else {
				break
			}
		}
	}

	// replace variables for "mapping"
	for k, v := range mapping {
		if old, replaced := replaceVar(&k); replaced {
			delete(mapping, old)
			mapping[k] = v
		}

		if _, replaced := replaceVar(&v); replaced {
			mapping[k] = v
		}
	}

	// replace glob for mapping

	// replace variables for parsedattrs
	for _, dirs := range parsedAttrsMap {
		for i := range dirs {
			replaceVar(&dirs[i])
		}
	}

	// replace glob for parsedattrs

}

func extractAll(syncdir string) map[string]string {
	res := make(map[string]string)

	allConfigFiles := lsDir(syncdir)

	for _, fname := range allConfigFiles {
		if filepath.Ext(fname) == attrFile {
			continue
		}
		f, err := os.Open(filepath.Join(syncdir, fname))
		if err != nil {
			log.Fatal(err)
		}
		scanner := bufio.NewScanner(f)
		curMode := mode{}

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, comment1) ||
				strings.HasPrefix(line, comment2) {
				continue
			}

			if line == "" {
				if curMode.inAttr {
					curMode.inAttr = false
					parsedAttrValueEachKey.Clear()
				}
				continue
			}

			// extract attributes
			matches := reAttr.FindStringSubmatch(line)
			if matches != nil {
				if curMode.inAttr {
					info := fmt.Sprintf("wrong reading config file, attr:%q\n"+
						"attributes should end with empty line\n", line)
					log.Fatalln(info)
				}

				if !supportedAttrs.Contains(matches[1]) {
					log.Fatalf("attribute %q not supported\n", matches[1])
				}
				curMode.inAttr = true
				curMode.name = matches[1]
				continue
			}

			if curMode.inAttr {
				if !paresdAttrKey.Contains(curMode.name) ||
					!parsedAttrValueEachKey.Contains(line) {
					parsedAttrsMap[curMode.name] = append(
						parsedAttrsMap[curMode.name], line)
					parsedAttrValueEachKey.Add(line)
					paresdAttrKey.Add(curMode.name)
				}
				continue
			}

			// extract variables
			if extractVariable(line) {
				continue
			}

			// extract directory map
			info := stringsW.SplitNoEmpty(line, mapSeparator)
			trimSpace(info)
			if len(info) == 0 {
				continue
			}
			if len(info) != 2 {
				fmt.Println("(shoud be) from : to", info)
				continue
			}
			k, v := info[0], info[1]
			if val, exist := res[k]; exist {
				log.Printf("key already set (%q:%q)\n", k, val)
			}

			if v == "" {
				log.Println("target file is empty")
				continue
			}
			res[k] = v
		}
	}
	return res
}

// change the global variable "variables"
func extractVariable(line string) bool {
	info := stringsW.SplitNoEmpty(line, "=")
	trimSpace(info)
	if len(info) == 0 {
		return false
	}

	if len(info) != 2 {
		return false
	}

	varName, varValue := info[0], info[1]
	if val, exist := variablesMap[varName]; exist {
		log.Printf("%q has already set to %q\n", varName, val)
	}

	variablesMap[varName] = varValue
	return true
}

// Parse parse all the files in "rootDir"
// returns parsed results
func Parse(rootDir string) *Result {
	initialize(rootDir)
	return &Result{
		Variables:  variablesMap,
		Mapping:    mapping,
		Attributes: parsedAttrsMap,
	}
}
