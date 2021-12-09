package _helpers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/utilsW"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const (
	fname   = ".go_tools_urls.txt"
	hintLen = 70
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

func PromptYesOrNo(msg string) bool {
	fmt.Print(msg)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	ans := strings.TrimSpace(scanner.Text())
	if strings.ToLower(ans) == "y" {
		return true
	}
	return false
}

func WriteUrls(titles []string) {
	if len(titles) < 1 {
		return
	}
	absName := filepath.Join(homeDir, fname)
	originalData := utilsW.ReadString(absName)
	f, err := os.OpenFile(absName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	matched := 0
	p := regexp.MustCompile(`(?i)^url[\s]*:`)
	for _, title := range titles {
		titleOneLine := strings.ReplaceAll(title, "\n", "")
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
			if p.MatchString(line) {
				matched++
				f.WriteString(p.ReplaceAllString(line, "") + "\x00" + titleOneLine)
				f.WriteString("\n")
			}
		}
	}
	if matched < 1 {
		f.WriteString(originalData)
	}

}

func OpenUrls() {
	absName := filepath.Join(homeDir, fname)
	f, err := os.Open(absName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
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
		fmt.Println(color.RedString("no urls are found"))
		return
	}
	if len(urls) == 1 {
		utilsW.OpenUrlInBrowswer(urls[0])
		return
	}
	// more than one urls
	urlsWithNo := make([]string, len(urls))
	for i := range urls {
		urlsWithNo[i] = fmt.Sprintf("%d: %s (%s)", i+1, color.GreenString(urls[i]),
			color.HiBlueString(hints[i]))
	}
	_print := func(urlsWithNo, info []string) {
		for i := range urlsWithNo {
			fmt.Println(urlsWithNo[i])
		}
	}
	_print(urlsWithNo, hints)
	fmt.Print("\ninput the number: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := strings.TrimSpace(scanner.Text())
	for {
		if val, err := strconv.Atoi(text); err != nil {
			_print(urlsWithNo, hints)
			fmt.Printf("%s is not a valid choice\n", text)
			scanner.Scan()
		} else if val > len(urls) {
			fmt.Printf("%d is too large\n", val)
		} else {
			utilsW.OpenUrlInBrowswer(urls[val-1])
			return
		}
	}
}
