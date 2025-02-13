package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/stringsW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	maxHistoryLines = 1
)

var (
	apiKey      string
	historyFile string
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
type RequestBody struct {
	Model        string    `json:"model"`
	Messages     []Message `json:"messages"`
	EnableSearch bool      `json:"enable_search"`
	Stream       bool      `json:"stream"`
}

func getText(j *utilsW.Json) string {
	choices := j.GetJson("choices")
	if choices.Len() == 0 {
		return ""
	}
	return choices.GetIndex(0).GetJson("delta").GetString("content")
}

func handleResponse(resp io.Reader) <-chan string {
	keyword := "data: {"
	ch := make(chan string)
	go func() {
		defer func() {
			close(ch)
			recover()
		}()
		for content := range stringsW.SplitByToken(resp, keyword, true) {
			if content == keyword || content == "" {
				continue
			}
			b := bytes.TrimRight(stringsW.StringToBytes(content)[len(keyword)-1:], "\n\t ")
			j := utilsW.NewJsonFromByte(b)
			ch <- getText(j)
		}
	}()
	return ch
}

func init() {
	config := utilsW.GetAllConfig()
	apiKey = config.GetOrDefault("api_key", "").(string)
	if apiKey == "" {
		fmt.Println("set api_key in ~/.configW")
		os.Exit(0)
	}

	historyFile = config.GetOrDefault("history_file", "").(string)
	if historyFile == "" {
		historyFile = utilsW.ExpandUser("~/.history_file.txt")
	}
}

func buildMessageArr(n int) []Message {
	if !utilsW.IsTextFile(historyFile) {
		return []Message{}
	}
	history := utilsW.ReadString(historyFile)
	result := make([]Message, 0)
	lines := stringsW.SplitNoEmptyKeepQuote(history, '\x01')
	for _, line := range lines {
		if line == "" {
			continue
		}
		arr := strings.Split(line, "\x00")
		role, content := arr[0], arr[1]
		result = append(result, Message{
			Role:    role,
			Content: content,
		})
	}
	if n > len(lines) {
		n = len(lines)
	}
	if len(lines) > maxHistoryLines {
		utilsW.WriteToFile(historyFile, stringsW.StringToBytes(strings.Join(lines[len(lines)-maxHistoryLines:], "\n")))
	}
	return result[len(lines)-n:]
}

func appendHistory(content string) {
	f, err := os.OpenFile(historyFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(content)
}

var sigChan = make(chan os.Signal, 1)

func exit() {
	// Create a channel to receive signals

	// Notify the sigChan channel for SIGINT (Ctrl+C) and SIGTERM signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// Block until a signal is received
	<-sigChan
}

func getQuestion(parsed *terminalW.ParsedResults) (question string) {
	multiLine := parsed.ContainsFlagStrict("multi-line") || parsed.ContainsFlagStrict("mul")
	if parsed.ContainsFlagStrict("e") {
		question = utilsW.InputWithEditor("", parsed.ContainsFlagStrict("code"))
	} else {
		question = utilsW.UserInput(color.GreenString("> "), multiLine)
	}
	return
}

func main() {
	go exit()
	flag.Int("history", 4, "number of history")
	flag.String("m", "", "model name. (qwen-max-latest(default), qwen-plus(default), qwen-max)")
	flag.Bool("h", false, "print help help")
	flag.Bool("multi-line", false, "input with multline")
	flag.Bool("mul", false, "same as multi-line")
	flag.Bool("e", false, "input with editor")
	flag.Bool("code", false, "input with vscode")
	parsed := terminalW.ParseArgsCmd("h", "multi-line", "mul", "e", "code")
	if parsed.ContainsFlagStrict("h") {
		flag.PrintDefaults()
		return
	}

	var nHistory = 4
	var model = parsed.GetFlagValueDefault("m", "qwen-plus")
	fmt.Println("Model: ", color.GreenString(model))
	var curr bytes.Buffer

	client := &http.Client{}
	for {
		question := getQuestion(parsed)
		curr.WriteString(fmt.Sprintf("%s\x00%s\x01", "user", question))

		// 构建请求体
		requestBody := RequestBody{
			// 模型列表：https://help.aliyun.com/zh/model-studio/getting-started/models
			Model: model,
			Messages: []Message{
				{
					Role:    "system",
					Content: "You are a helpful assistant.",
				},
			},
			EnableSearch: true,
			Stream:       true,
		}
		arr := buildMessageArr(nHistory)
		requestBody.Messages = append(requestBody.Messages, arr...)
		requestBody.Messages = append(requestBody.Messages, Message{
			Role:    "user",
			Content: question,
		})
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			log.Fatal(err)
		}
		// 创建 POST 请求
		req, err := http.NewRequest("POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		curr.WriteString("assistant\x00")
		ch := handleResponse(resp.Body)

		for {
			select {
			case <-sigChan:
				goto end
			case content, ok := <-ch:
				if !ok {
					goto end
				}
				curr.WriteString(content)
				fmt.Print(content)
			default:
				time.Sleep(1000)
			}
		}
	end:
		resp.Body.Close()
		curr.WriteByte('\x01')
		appendHistory(curr.String())
		if parsed.ContainsAnyFlagStrict("e", "code") {
			if !utilsW.PromptYesOrNo("\ncontinue? (y/n)") {
				return
			}
		}
		fmt.Println()
	}
}
