package _helpers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/containerW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const (
	urlFileName    = ".go_tools_urls.txt"
	commonFileName = ".go_tools_common.txt"
	opFileName     = ".go_tools_previous_op.txt"
	hintLen        = 120
)

var (
	homeDir string
)

func init() {
	var err error
	homeDir, err = os.UserHomeDir()
	if err != nil {
		panic(err)
	}
}

func _print(urlsWithNo, info []string) {
	for i := range urlsWithNo {
		fmt.Println(urlsWithNo[i])
	}
}

func CollectionExists(db *mongo.Database, ctx context.Context, collectionName string) bool {
	names, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		panic(err)
	}
	for _, name := range names {
		if name == collectionName {
			return true
		}
	}
	return false
}

func WritePreviousOpration(op string) {
	if err := ioutil.WriteFile(filepath.Join(homeDir, opFileName), []byte(op), 0666); err != nil {
		panic(err)
	}
}

func WriteInfo(objectIDs []*primitive.ObjectID, titles []string) bool {
	if len(titles) < 1 {
		return false
	}
	if len(objectIDs) != len(titles) {
		log.Println("objectIDs length doesn't equal to titles length")
		return false
	}
	absName := filepath.Join(homeDir, urlFileName)
	absNameCommon := filepath.Join(homeDir, commonFileName)
	var originalData, originalCommonData string
	if utilsW.IsExist(absName) {
		originalData = utilsW.ReadString(absName)
	}
	if utilsW.IsExist(absNameCommon) {
		originalCommonData = utilsW.ReadString(absNameCommon)
	}
	f, err := os.OpenFile(absName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	commonF, err := os.OpenFile(absNameCommon, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer commonF.Close()

	matched := 0
	p := regexp.MustCompile(`(?i)url[\s]*[:：]`)
	for i := range titles {
		title := titles[i]
		objectID := objectIDs[i]
		titleOneLine := strings.ReplaceAll(title, "\n", "")
		titleOneLine = strings.ReplaceAll(titleOneLine, "\r", "")
		buf := bytes.NewBufferString("")
		for i, ch := range titleOneLine {
			if i >= hintLen {
				break
			}
			buf.WriteRune(ch)
		}
		titleOneLine = buf.String()
		buf = bytes.NewBufferString(title)
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			line = strings.ReplaceAll(line, "\x00", "")
			// write urls
			// only when matched
			if p.MatchString(line) {
				matched++
				f.WriteString(p.ReplaceAllString(line, "") + "\x00" + titleOneLine)
				f.WriteString("\n")
			}
		}
		// write information all the time
		commonF.WriteString(objectID.Hex() + "\x00" + titleOneLine)
		commonF.WriteString("\n")
	}
	written := true
	if matched < 1 {
		written = false
	}
	if matched < 1 && originalData != "" {
		f.WriteString(originalData)
	}
	if len(titles) < 1 && originalCommonData != "" {
		commonF.WriteString(originalCommonData)
	}
	return written
}

// ReadInfo   如果isURL=true，返回空字符串
// 如果isURL=false，返回查询到的 ObjectID
func ReadInfo(isURL bool) string {
	fname := urlFileName
	if !isURL {
		fname = commonFileName
	}
	absName := filepath.Join(homeDir, fname)
	f, err := os.Open(absName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	// urls may not be url, they can be ObjectIDs
	urls := make([]string, 0)
	hints := make([]string, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		info := strings.Split(line, "\x00")
		url, hint := info[0], info[1]
		urls = append(urls, url)
		hints = append(hints, hint)
	}

	if len(urls) == 0 {
		msg := "urls"
		if !isURL {
			msg = "ObjectIDs"
		}
		fmt.Println(color.RedString(fmt.Sprintf("no %s are found", msg)))
		return ""
	}
	if len(urls) == 1 {
		if isURL {
			utilsW.OpenUrlInBrowswer(urls[0])
			return ""
		}
		return urls[0]
	}
	// more than one urls
	urlsWithNo := make([]string, len(urls))
	for i := range urls {
		urlsWithNo[i] = fmt.Sprintf("%d: %s (%s)", i+1, color.HiWhiteString(urls[i]),
			color.HiBlueString(hints[i]))
	}
	_print(urlsWithNo, hints)
	fmt.Print("\ninput the number: ")
	scanner = bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if val, err := strconv.Atoi(text); err != nil {
			fmt.Printf("%s is not a valid choice\n", text)
			_print(urlsWithNo, hints)
			fmt.Print("\ninput the number: ")
		} else if val > len(urls) {
			fmt.Printf("%d is too large\n", val)
			_print(urlsWithNo, hints)
			fmt.Print("\ninput the number: ")
		} else {
			if isURL {
				utilsW.OpenUrlInBrowswer(urls[val-1])
				return ""
			} else {
				return urls[val-1]
			}
		}
	}
	return ""
}

func OrderByTime(parsed *terminalW.ParsedResults) bool {
	if parsed == nil {
		return false
	}
	if parsed.ContainsFlagStrict("t") && parsed.MustGetFlagVal("t") == "" {
		return true
	}
	if parsed.ContainsAnyFlagStrict("ti", "it") {
		return true
	}
	if parsed.Positional.Contains("it") || parsed.Positional.Contains("ti") {
		return true
	}
	return false
}

// IsObjectID 返回是否是有效的mongodb objectid
func IsObjectID(id string) bool {
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return false
	}
	return true
}

func BuildMongoRegularExpExclude(specialPattern *containerW.Set) string {
	if specialPattern.Size() == 1 {
		return fmt.Sprintf("^(?!%s).*", specialPattern.ToSlice()[0].(string))
	}
	res := bytes.NewBufferString("^(?!(")
	for val := range specialPattern.Iterate() {
		res.WriteString(val.(string))
		res.WriteString("|")
	}
	res.Truncate(res.Len() - 1)
	res.WriteString(")).*")
	return res.String()
}

func SearchTrie(trie *containerW.Trie, specialPattern *containerW.Set) bool {
	for val := range specialPattern.Iterate() {
		if trie.StartsWith(val.(string)) {
			return true
		}
	}
	return false
}
