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
	fname = ".go_tools_urls.txt"
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
	absName := filepath.Join(homeDir, fname)
	f, err := os.OpenFile(absName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	p := regexp.MustCompile(`(?i)^url[\s]*:`)
	for _, title := range titles {
		buf := bytes.NewBufferString(title)
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimSpace(line)
			if p.MatchString(line) {
				f.WriteString(p.ReplaceAllString(line, ""))
				f.WriteString("\n")
			}
		}
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
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		urls = append(urls, url)
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
		urlsWithNo[i] = fmt.Sprintf("%d: %s", i+1, color.GreenString(urls[i]))
	}
	_print := func(urlsWithNo []string) {
		for _, line := range urlsWithNo {
			fmt.Println(line)
		}
	}
	_print(urlsWithNo)
	fmt.Print("input the number: ")
	scanner = bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := strings.TrimSpace(scanner.Text())
	for {
		if val, err := strconv.Atoi(text); err != nil {
			_print(urlsWithNo)
			fmt.Printf("%s is not a valid choice\n")
			scanner.Scan()
		} else {
			utilsW.OpenUrlInBrowswer(urls[val-1])
			return
		}
	}
}
