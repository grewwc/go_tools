package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/executable/ai/_ai_helpers"
	"github.com/grewwc/go_tools/src/strW"
	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	maxHistoryLines   = 100
	defaultNumHistory = 4
)

var (
	apiKey      string
	historyFile string
)

var (
	fileidArr []string
)

var (
	thinking int32 = 0
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

var (
	nHistory int
)

var (
	thinkingTag    = color.YellowString("<thinking>")
	endThinkingTag = color.YellowString("<end thinking>")
)

func getText(j *utilsW.Json) string {
	choices := j.GetJson("choices")
	if choices.Len() == 0 {
		return ""
	}
	delta := choices.GetIndex(0).GetJson("delta")
	content := delta.GetString("content")
	if content == "" && delta.ContainsKey("reasoning_content") {
		// thinking
		content := delta.GetString("reasoning_content")
		if content != "" && atomic.LoadInt32(&thinking) == 0 {
			content = fmt.Sprintf("\n%s\n%s", thinkingTag, content)
			atomic.AddInt32(&thinking, 1)
		}
		return content
	} else {
		if atomic.LoadInt32(&thinking) == 1 {
			content = fmt.Sprintf("\n%s\n%s", endThinkingTag, content)
			atomic.AddInt32(&thinking, -1)
		}
	}
	return content
}

func handleResponse(resp io.Reader) <-chan string {
	keyword := "data: {"
	doneKeyword := strW.StringToBytes("data: [DONE]")
	ch := make(chan string)
	go func() {
		defer func() {
			close(ch)
			if err := recover(); err != nil {
				log.Fatalln(err)
			}
		}()
		for content := range strW.SplitByToken(resp, keyword, true) {
			if content == keyword || content == "" {
				continue
			}
			b := bytes.TrimRight(strW.StringToBytes(content)[len(keyword)-1:], "\n\t ")
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
	lines := strW.SplitNoEmptyKeepQuote(history, '\x01')
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
		utilsW.WriteToFile(historyFile, strW.StringToBytes(strings.Join(lines[len(lines)-maxHistoryLines:], "\n")))
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

func modifyQuestion(question string) string {
	if strings.HasSuffix(strings.TrimSpace(question), " -s") {
		question = strings.TrimSuffix(strings.TrimSpace(question), " -s")
		question += "\n Please be concise."
	}
	return question
}

func getQuestion(parsed *terminalW.Parser, userInput bool) (question string) {
	var fileContent string
	if userInput {
		multiLine := parsed.ContainsFlagStrict("multi-line") || parsed.ContainsFlagStrict("mul")
		question = utilsW.UserInput("> ", multiLine)
		tempParser := terminalW.NewParser()
		tempParser.Bool("x", false, "")
		tempParser.Bool("c", false, "")
		tempParser.ParseArgs(fmt.Sprintf("a %s", question), "x", "c")
		if tempParser.GetFlagValueDefault("f", "") != "" {
			parsed.SetFlagValue("f", tempParser.GetFlagValueDefault("f", ""))
		}
		if tempParser.ContainsFlagStrict("c") {
			parsed.SetFlagValue("c", "true")
		}
		if tempParser.ContainsFlagStrict("s") {
			parsed.SetFlagValue("s", "true")
		}
		question = strings.Join(tempParser.GetPositionalArgs(true), " ")
		if tempParser.GetNumArgs() != -1 {
			question = fmt.Sprintf("%s -%d", question, tempParser.GetNumArgs())
		}
		nHistory = getNumHistory(tempParser)
	} else {
		question = strings.Join(parsed.GetPositionalArgs(true), " ")
		nHistory = getNumHistory(parsed)
	}
	if parsed.GetFlagValueDefault("f", "") != "" {
		files := parsed.MustGetFlagVal("f")
		parser := _ai_helpers.NewParser(files)
		question = parser.TextFileContents() + "\n" + question
		_ai_helpers.NonTextFile.Set(parser.NonTextFiles())
		parsed.RemoveFlagValue("f")
	}
	if parsed.ContainsFlagStrict("c") {
		fileContent += utilsW.ReadClipboardText()
		parsed.RemoveFlagValue("c")
	}
	// short output
	if parsed.ContainsFlagStrict("s") {
		question += "\n Please be concise."
		parsed.RemoveFlagValue("s")
	}
	question = fileContent + question
	return
}

func getNumHistory(parsed *terminalW.Parser) int {
	if parsed.ContainsFlagStrict("x") {
		return 0
	}
	return parsed.GetIntFlagValOrDefault("history", defaultNumHistory)
}

func getWriteResultFile(parsed *terminalW.Parser) *os.File {
	if parsed.ContainsFlagStrict("out") {
		filename := parsed.GetFlagValueDefault("out", "")
		if filename == "" {
			filename = "output.txt"
		}
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		return f
	} else {
		return nil
	}
}

func main() {
	// Notify the sigChan channel for SIGINT (Ctrl+C) and SIGTERM signals
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sigCh1 := make(chan os.Signal, 1)
	signal.Notify(sigCh1, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for range sigCh1 {
			fmt.Println("Exit.")
			<-sigChan
			os.Exit(0)
		}
	}()

	parser := terminalW.NewParser()
	parser.Int("history", defaultNumHistory, "number of history")
	parser.String("m", "", `model name. (configured by \"ai.model.default\") ,
qwq-plus[0], qwen-plus[1], qwen-max[2], qwen-max-latest[3], qwen-coder-plus-latest [4], deepseek-r1 [5], qwen-turbo [6, default])`)
	parser.Bool("h", false, "print help info")
	parser.Bool("multi-line", false, "input with multline")
	parser.Bool("mul", false, "same as multi-line")
	parser.Bool("code", false, "use code model (qwen-coder-plus-latest)")
	parser.Bool("s", false, "short output")
	parser.Bool("d", false, "deepseek model")
	parser.Bool("clear", false, "clear history")
	parser.Bool("c", false, "prepend content in clipboard")
	parser.Bool("x", false, "ask with history")
	parser.String("f", "", "input file names. seprated by comma.")
	parser.String("out", "", "write output to file. default is output.txt")
	parser.ParseArgsCmd("h", "multi-line", "mul", "code", "s", "d", "-clear", "c", "x")
	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}

	if parser.ContainsFlagStrict("-clear") {
		_ = os.Remove(historyFile)
		return
	}

	args := parser.Positional.ToStringSlice()

	var model = _ai_helpers.GetModel(parser)
	var curr bytes.Buffer

	client := &http.Client{}
	var f *os.File = getWriteResultFile(parser)
	var out io.Writer = os.Stdout
	var shouldQuit bool
	if f != nil {
		defer f.Close()
		out = io.MultiWriter(out, f)
	}
	for {
		var question string
		if len(args) >= 1 {
			question = getQuestion(parser, false)
			args = []string{}
			shouldQuit = true
		} else {
			question = getQuestion(parser, true)
		}
		if strings.TrimSpace(question) == "" {
			continue
		}
		curr.WriteString(fmt.Sprintf("%s\x00%s\x01", "user", question))

		nextModel := _ai_helpers.GetModelByInput(model, &question)
		model = nextModel
		question = modifyQuestion(question)
		// 构建请求体
		requestBody := RequestBody{
			// 模型列表：https://help.aliyun.com/zh/model-studio/getting-started/models
			Model: nextModel,
			Messages: []Message{
				{
					Role:    "system",
					Content: "You are a helpful assistant.",
				},
			},
			EnableSearch: _ai_helpers.SearchEnabled(nextModel),
			Stream:       true,
		}
		files := _ai_helpers.NonTextFile.Get().([]string)
		if len(files) > 0 || len(fileidArr) > 0 {
			_ai_helpers.NonTextFile.Set([]string{})
			if len(files) > 0 {
				fileidArr = _ai_helpers.UploadQwenLongFiles(apiKey, files)
			}
			fileids := strings.Join(fileidArr, ",")
			msg := Message{
				Role:    "system",
				Content: fmt.Sprintf("fileid://%s", fileids),
			}
			requestBody.Messages = append(requestBody.Messages, msg)
			requestBody.Model = _ai_helpers.QWEN_LONG
		} else {
			arr := buildMessageArr(nHistory)
			requestBody.Messages = append(requestBody.Messages, arr...)
		}
		requestBody.Messages = append(requestBody.Messages, Message{
			Role:    "user",
			Content: question,
		})
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			log.Fatalln(err)
		}
		// 创建 POST 请求
		req, err := http.NewRequest("POST", _ai_helpers.GetEndpoint(), bytes.NewBuffer(jsonData))
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		// 发送请求
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalln(err)
		}
		curr.WriteString("assistant\x00")
		ch := handleResponse(resp.Body)

		fmt.Printf("[%s] ", color.GreenString(requestBody.Model))
		for {
			select {
			case <-sigChan:
				goto end
			case content, ok := <-ch:
				if !ok {
					goto end
				}
				fmt.Fprint(out, content)
				if atomic.LoadInt32(&thinking) == 1 {
					continue
				}
				content = strings.Replace(content, endThinkingTag, "", 1)
				content = strings.Trim(content, "\n")
				curr.WriteString(content)
				time.Sleep(0)
			}
		}
	end:
		resp.Body.Close()
		curr.WriteByte('\x01')
		appendHistory(curr.String())
		fmt.Println()
		if shouldQuit {
			return
		}
	}
}
