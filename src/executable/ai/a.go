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
	internal "github.com/grewwc/go_tools/src/executable/ai/internel"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw"
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

	raw bool = false
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

func getText(j *utilsw.Json) string {
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
	doneKeyword := typesw.StrToBytes("data: [DONE]")
	ch := make(chan string)
	go func() {
		defer func() {
			recover()
			close(ch)
		}()
		for content := range strw.SplitByToken(resp, keyword, true) {
			if content == keyword || content == "" {
				continue
			}
			b := typesw.StrToBytes(content)
			b = bytes.TrimSpace(b)
			b = bytes.TrimSuffix(b, doneKeyword)
			b = bytes.TrimSuffix(b, typesw.StrToBytes(keyword))
			b = bytes.TrimSpace(b)
			b = append([]byte{'{'}, b...)
			// fmt.Println("==>")
			// fmt.Println(string(b))
			// fmt.Println("===")
			// os.Exit(0)
			j, err := utilsw.NewJsonFromByte(b)
			if err != nil {
				log.Println("handleResponse error", err)
				fmt.Println("======> response: ")
				fmt.Println(string(b))
				fmt.Println("<======")
			}
			ch <- getText(j)
		}
	}()
	return ch
}

func init() {
	config := utilsw.GetAllConfig()
	apiKey = config.GetOrDefault("api_key", "").(string)
	if apiKey == "" {
		fmt.Println("set api_key in ~/.configW")
		os.Exit(0)
	}

	historyFile = config.GetOrDefault("history_file", "").(string)
	if historyFile == "" {
		historyFile = utilsw.ExpandUser("~/.history_file.txt")
	}
}

func buildMessageArr(n int) []Message {
	if !utilsw.IsTextFile(historyFile) {
		return []Message{}
	}
	history := utilsw.ReadString(historyFile)
	result := make([]Message, 0)
	lines := strw.SplitNoEmptyPreserveQuote(history, '\x01', `"`, true)
	for _, line := range lines {
		if line == "" {
			continue
		}
		arr := strings.Split(line, "\x00")
		if len(arr) < 2 {
			continue
		}
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
		utilsw.WriteToFile(historyFile, typesw.StrToBytes(strings.Join(lines[len(lines)-maxHistoryLines:], "\n")))
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

func getQuestion(parsed *terminalw.Parser, loopMode bool) (question string) {
	var fileContent string
	if loopMode {
		multiLine := parsed.ContainsFlagStrict("multi-line") || parsed.ContainsFlagStrict("mul")
		question = utilsw.UserInput("> ", multiLine)
		tempParser := terminalw.NewParser(terminalw.DisableParserNumber)
		// tempParser.Bool("x", false, "")
		// tempParser.Bool("c", false, "")
		tempParser.ParseArgs(question)
		// if tempParser.GetFlagValueDefault("f", "") != "" {
		// 	parsed.SetFlagValue("f", tempParser.GetFlagValueDefault("f", ""))
		// }
		// if tempParser.ContainsFlagStrict("c") {
		// 	parsed.SetFlagValue("c", "true")
		// }
		// if tempParser.ContainsFlagStrict("s") {
		// 	parsed.SetFlagValue("s", "true")
		// }
		question = strings.Join(tempParser.GetPositionalArgs(true), " ")
		// fmt.Println("getQuestion: ")
		// fmt.Println(question)
		// os.Exit(0)
		// if tempParser.GetNumArgs() != -1 {
		// 	question = fmt.Sprintf("%s -%d", question, tempParser.GetNumArgs())
		// }
		nHistory = getNumHistory(tempParser)
	} else {
		if raw {
			question = strings.Join(os.Args[1:], " ")
		} else {
			question = strings.Join(parsed.GetPositionalArgs(true), " ")
		}

		nHistory = getNumHistory(parsed)
	}
	if parsed.GetFlagValueDefault("f", "") != "" {
		files := parsed.MustGetFlagVal("f")
		parser := internal.NewParser(files)
		question = parser.TextFileContents() + "\n" + question
		internal.NonTextFile.Set(parser.NonTextFiles())
		parsed.RemoveFlagValue("f")
	}
	if parsed.ContainsFlagStrict("c") {
		fileContent += utilsw.ReadClipboardText()
		parsed.RemoveFlagValue("c")
	}
	// short output
	question = fileContent + question
	return
}

func getNumHistory(parsed *terminalw.Parser) int {
	if parsed.ContainsFlagStrict("x") {
		return 0
	}
	return parsed.GetIntFlagValOrDefault("history", defaultNumHistory)
}

func getWriteResultFile(parsed *terminalw.Parser) *os.File {
	if parsed.ContainsFlagStrict("out") {
		filename := parsed.GetFlagValueDefault("out", "")
		if filename == "" {
			filename = "output.md"
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

	parser := terminalw.NewParser()
	parser.Int("history", defaultNumHistory, "number of history")
	parser.String("m", "", `model name. (configured by \"ai.model.default\") ,
qwq-plus[0], qwen-plus[1], qwen-max[2], qwen-max-latest[3], qwen-coder-plus-latest [4], deepseek-r1 [5], qwen-turbo [6]`)
	parser.Bool("h", false, "print help info")
	parser.Bool("multi-line", false, "input with multline")
	parser.Bool("mul", false, "same as multi-line")
	parser.Bool("code", false, "use code model (qwen-coder-plus-latest)")
	parser.Bool("d", false, "deepseek model")
	parser.Bool("clear", false, "clear history")
	parser.Bool("c", false, "prepend content in clipboard")
	parser.Bool("x", false, "ask without history")
	parser.String("f", "", "input file names. seprated by comma.")
	parser.String("out", "", "write output to file. default is output.txt")
	parser.Bool("raw", false, "raw mode: don't use parser to get positional arguments, use raw inputs instead.")
	parser.ParseArgsCmd()
	if parser.ContainsFlagStrict("h") {
		parser.PrintDefaults()
		return
	}

	if parser.ContainsFlagStrict("-clear") {
		os.Remove(historyFile)
		fmt.Println("History cleared.")
		return
	}

	raw = parser.ContainsFlagStrict("raw")

	args := parser.GetPositionalArgs(true)
	// args := os.Args[1:]

	var model = internal.GetModel(parser)
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

		nextModel := internal.GetModelByInput(model, &question)
		model = nextModel
		// fmt.Println("here question:", question)
		// 构建请求体
		// fmt.Println(internal.SearchEnabled(model), model)
		requestBody := RequestBody{
			// 模型列表：https://help.aliyun.com/zh/model-studio/getting-started/models
			Model: nextModel,
			Messages: []Message{
				{
					Role:    "system",
					Content: "You are a helpful assistant.",
				},
			},
			EnableSearch: internal.SearchEnabled(nextModel),
			Stream:       true,
		}
		files := internal.NonTextFile.Get().([]string)
		if len(files) > 0 || len(fileidArr) > 0 {
			internal.NonTextFile.Set([]string{})
			if len(files) > 0 {
				fileidArr = internal.UploadQwenLongFiles(apiKey, files)
			}
			fileids := strings.Join(fileidArr, ",")
			msg := Message{
				Role:    "system",
				Content: fmt.Sprintf("fileid://%s", fileids),
			}
			requestBody.Messages = append(requestBody.Messages, msg)
			requestBody.Model = internal.QWEN_LONG
		} else {
			arr := buildMessageArr(nHistory)
			requestBody.Messages = append(requestBody.Messages, arr...)
		}
		requestBody.Messages = append(requestBody.Messages, Message{
			Role:    "user",
			Content: question,
		})
		jsonData, _ := json.Marshal(requestBody)
		req, _ := http.NewRequest("POST", internal.GetEndpoint(), bytes.NewBuffer(jsonData))
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", "application/json")
		// 发送请求
		resp, _ := client.Do(req)
		curr.WriteString("assistant\x00")
		ch := handleResponse(resp.Body)
		search := "true"
		if !internal.SearchEnabled(model) {
			search = "false"
		}
		fmt.Printf("[%s (search: %s)] ", color.GreenString(requestBody.Model), color.RedString(search))
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
		f.WriteString("\n---\n")
	}
}
