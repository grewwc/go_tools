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
	maxHistoryLines   = 100
	defaultNumHistory = 4
)

const (
	QWEN_MAX_LASTEST       = "qwen-max-latest"
	QWEN_PLUS              = "qwen-plus"
	QWEN_MAX               = "qwen-max"
	QWEN_CODER_PLUS_LATEST = "qwen-coder-plus-latest"
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
	doneKeyword := stringsW.StringToBytes("data: [DONE]")
	ch := make(chan string)
	go func() {
		defer func() {
			close(ch)
			if err := recover(); err != nil {
				log.Fatalln(err)
			}
		}()
		for content := range stringsW.SplitByToken(resp, keyword, true) {
			if content == keyword || content == "" {
				continue
			}
			b := bytes.TrimRight(stringsW.StringToBytes(content)[len(keyword)-1:], "\n\t ")
			b = bytes.TrimSuffix(b, doneKeyword)
			b = bytes.TrimSpace(b)
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

func getQuestion(parsed *terminalW.ParsedResults) (question string) {
	multiLine := parsed.ContainsFlagStrict("multi-line") || parsed.ContainsFlagStrict("mul")
	if parsed.ContainsFlagStrict("e") {
		question = utilsW.InputWithEditor("", parsed.ContainsFlagStrict("vs"))
	} else {
		question = utilsW.UserInput(color.GreenString("> "), multiLine)
	}
	return
}

func getModel(parsed *terminalW.ParsedResults) string {
	if parsed.ContainsFlagStrict("code") {
		return QWEN_CODER_PLUS_LATEST
	}

	model := parsed.GetFlagValueDefault("m", QWEN_PLUS)
	switch model {
	case QWEN_PLUS, "1":
		return QWEN_PLUS
	case QWEN_MAX, "2":
		return QWEN_MAX
	case QWEN_MAX_LASTEST, "3":
		return QWEN_MAX_LASTEST
	case QWEN_CODER_PLUS_LATEST, "4":
		return QWEN_CODER_PLUS_LATEST
	default:
		return QWEN_PLUS
	}
}

func getNumHistory(parsed *terminalW.ParsedResults) int {
	res := parsed.GetNumArgs()
	if res != -1 {
		return res
	}
	return defaultNumHistory
}

func getWriteResultFile(parsed *terminalW.ParsedResults) *os.File {
	if parsed.ContainsFlagStrict("f") {
		filename := parsed.GetFlagValueDefault("f", "")
		if filename == "" {
			filename = "output.txt"
		}
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		return f
	} else {
		return nil
	}
}

func writeToFile(f *os.File, content string) {
	if f == nil {
		return
	}
	if _, err := io.WriteString(f, content); err != nil {
		panic(err)
	}
}

func main() {
	// Notify the sigChan channel for SIGINT (Ctrl+C) and SIGTERM signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	stopChan := make(chan struct{})
	go func() {
		<-sigChan
		close(stopChan)
	}()
	flag.Int("history", defaultNumHistory, fmt.Sprintf("number of history (default: %d)", defaultNumHistory))
	flag.String("m", "", "model name. (qwen-plus[1, default], qwen-max[2], qwen-max-latest[3], qwen-coder-plus-latest [4])")
	flag.Bool("h", false, "print help help")
	flag.Bool("multi-line", false, "input with multline")
	flag.Bool("mul", false, "same as multi-line")
	flag.Bool("e", false, "input with editor")
	flag.Bool("vs", false, "input with vscode")
	flag.Bool("code", false, "use code model (qwen-coder-plus-latest)")
	flag.String("f", "", "write output to file")
	parsed := terminalW.ParseArgsCmd("h", "multi-line", "mul", "e", "code", "vs")
	if parsed.ContainsFlagStrict("h") {
		flag.PrintDefaults()
		return
	}

	var nHistory = getNumHistory(parsed)
	var model = getModel(parsed)
	fmt.Println("Model: ", color.GreenString(model))
	var curr bytes.Buffer

	client := &http.Client{}
	var f *os.File = getWriteResultFile(parsed)
	if f != nil {
		defer f.Close()
	}
	for {
		question := getQuestion(parsed)
		if strings.TrimSpace(question) == "" {
			continue
		}
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
			EnableSearch: model != QWEN_CODER_PLUS_LATEST,
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
			case <-stopChan:
				goto end
			case content, ok := <-ch:
				if !ok {
					goto end
				}
				curr.WriteString(content)
				fmt.Print(content)
				writeToFile(f, content)
				time.Sleep(0)
			}
		}
	end:
		resp.Body.Close()
		curr.WriteByte('\x01')
		appendHistory(curr.String())
		if parsed.ContainsAnyFlagStrict("e", "vs") {
			if !utilsW.PromptYesOrNo("\ncontinue? (y/n)") {
				return
			}
		}
		fmt.Println()
	}
}
