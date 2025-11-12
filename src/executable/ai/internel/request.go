package internal

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/typesw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	MaxHistoryLines   = 100
	DefaultNumHistory = 4
)

const (
	Colon   = '\x00'
	Newline = '\x01'
)

type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}
type RequestBody struct {
	Model        string    `json:"model"`
	Messages     []Message `json:"messages"`
	EnableSearch bool      `json:"enable_search"`
	Stream       bool      `json:"stream"`
}

func buildMessageArr(n int, historyFile string) []Message {
	if !utilsw.IsTextFile(historyFile) {
		return []Message{}
	}
	history := utilsw.ReadString(historyFile)
	result := make([]Message, 0)
	lines := strw.SplitByStrKeepQuotes(history, string(Newline), `"`, false)
	validLines := 0

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Find the last occurrence of the colon separator
		lastColon := strings.LastIndex(line, string(Colon))
		if lastColon <= 0 || lastColon >= len(line)-1 {
			continue
		}

		role := line[:lastColon]
		content := line[lastColon+1:]

		// Skip invalid roles
		if role != "user" && role != "assistant" {
			continue
		}

		result = append(result, Message{
			Role:    role,
			Content: content,
		})
		validLines++
	}

	if n > validLines {
		n = validLines
	}

	// Trim history file if it's too long
	if len(lines) > MaxHistoryLines {
		utilsw.WriteToFile(historyFile, typesw.StrToBytes(strings.Join(lines[len(lines)-MaxHistoryLines:], string(Newline))))
	}

	// Return the last n messages
	if validLines > n {
		return result[validLines-n:]
	}
	return result
}

func getClient() *http.Client {
	return utilsw.Call[*http.Client](1, nil, func(a ...any) any {
		client := &http.Client{}
		return client
	})
}

func printInfo(model string) {
	search := "true"
	if !searchEnabled(model) {
		search = "false"
	}
	fmt.Printf("[%s (search: %s)] ", color.GreenString(model), color.RedString(search))
}

func buildContent(model, question string, fname []string) any {
	if !isVlModel(model) {
		return question
	}
	if len(fname) == 0 {
		return question
	}
	j := utilsw.NewJson(nil)
	for _, name := range fname {
		f, err := os.Open(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open file: %q", fname)
			return question
		}
		defer f.Close()
		b, err := io.ReadAll(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to readfile file: %q", fname)
			return question
		}
		image := base64.StdEncoding.EncodeToString(b)
		image = "data:image/jpg;base64," + image
		imageJ := utilsw.NewJson(nil)
		imageJ.Set("image_url", image)
		imageJ.Set("type", "image_url")
		j.Add(imageJ)
	}
	textJ := utilsw.NewJson(nil)
	textJ.Set("text", question)
	textJ.Set("type", "text")
	j.Add(textJ)
	return j.RawData()
}

func DoRequest(apiKey, model, question, historyFile string) *http.Response {
	requestBody := RequestBody{
		// 模型列表：https://help.aliyun.com/zh/model-studio/getting-started/models
		Model: model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful assistant.",
			},
		},
		EnableSearch: searchEnabled(model),
		Stream:       true,
	}

	files := NonTextFile.Get().([]string)
	if (len(files) > 0 || len(fileidArr) > 0) && !(isVlModel(model)) {
		NonTextFile.Set([]string{})
		if len(files) > 0 {
			fileidArr = uploadQwenLongFiles(apiKey, files)
		}
		fileids := strings.Join(fileidArr, ",")
		msg := Message{
			Role:    "system",
			Content: fmt.Sprintf("fileid://%s", fileids),
		}
		requestBody.Messages = append(requestBody.Messages, msg)
		requestBody.Model = QWEN_LONG
	} else if !isVlModel(model) {
		arr := buildMessageArr(NHistory, historyFile)
		requestBody.Messages = append(requestBody.Messages, arr...)
	}

	requestBody.Messages = append(requestBody.Messages, Message{
		Role:    "user",
		Content: buildContent(model, question, files),
	})

	jsonData, _ := json.Marshal(requestBody)
	req, _ := http.NewRequest("POST", getEndpoint(), bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	client := getClient()
	printInfo(model)
	utilsw.NewJson(requestBody).ToFile("test.json")
	// os.Exit(0)
	resp, _ := client.Do(req)
	return resp
}
