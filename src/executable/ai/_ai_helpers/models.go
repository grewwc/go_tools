package _ai_helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	DEEPSEEK               = "deepseek-r1-0528"
	QWEN_MAX_LASTEST       = "qwen-max-latest"
	QWEN_PLUS              = "qwen-plus"
	QWEN_MAX               = "qwen-max"
	QWEN_CODER_PLUS_LATEST = "qwen-coder-plus-latest"
	QWEN_LONG              = "qwen-long"
	QWQ                    = "qwq-plus-latest"
	QWEN_TURBO_LATEST      = "qwen-turbo-latest"
)

const (
	QWEN_ENDPOINT = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
)

func getDefaultModel() string {
	config := utilsw.GetAllConfig()
	return config.GetOrDefault("ai.model.default", QWEN_TURBO_LATEST).(string)
}

func GetEndpoint() string {
	config := utilsw.GetAllConfig()
	return config.GetOrDefault("ai.model.endpoint", QWEN_ENDPOINT).(string)
}

func GetModel(parsed *terminalw.Parser) string {
	if parsed.ContainsFlagStrict("code") {
		return QWEN_CODER_PLUS_LATEST
	}
	if parsed.ContainsFlagStrict("d") {
		return DEEPSEEK
	}
	n := parsed.GetNumArgs()
	switch n {
	case 0:
		return QWQ
	case 1:
		return QWEN_PLUS
	case 2:
		return QWEN_MAX
	case 3:
		return QWEN_MAX_LASTEST
	case 4:
		return QWEN_CODER_PLUS_LATEST
	case 5:
		return DEEPSEEK
	}
	model := parsed.GetFlagValueDefault("m", getDefaultModel())

	switch model {
	case QWQ, "0":
		return QWQ
	case QWEN_PLUS, "1":
		return QWEN_PLUS
	case QWEN_MAX, "2":
		return QWEN_MAX
	case QWEN_MAX_LASTEST, "3":
		return QWEN_MAX_LASTEST
	case QWEN_CODER_PLUS_LATEST, "4":
		return QWEN_CODER_PLUS_LATEST
	case DEEPSEEK, "5":
		return DEEPSEEK
	default:
		return getDefaultModel()
	}
}

var NonTextFile = utilsw.NewThreadSafeVal([]string{})

func GetModelByInput(prevModel string, input *string) string {
	if len(NonTextFile.Get().([]string)) > 0 {
		return QWEN_LONG
	}
	if prevModel == QWEN_LONG {
		return QWEN_LONG
	}
	trimed := strings.TrimSpace(*input)
	if strings.HasSuffix(trimed, " -code") {
		*input = strings.TrimSuffix(trimed, " -code")
		return QWEN_CODER_PLUS_LATEST
	}
	if strings.HasSuffix(trimed, " -d") {
		*input = strings.TrimSuffix(trimed, " -d")
		return DEEPSEEK
	}

	p := regexp.MustCompile(` -\d$`)
	if found := p.FindString(trimed); found != "" {
		*input = p.ReplaceAllString(trimed, "")
		parser := terminalw.NewParser()
		parser.ParseArgs(fmt.Sprintf("a %s", found))
		return GetModel(parser)
	}

	return prevModel
}

func SearchEnabled(model string) bool {
	return model == QWEN_MAX || model == QWEN_MAX_LASTEST || model == QWEN_PLUS ||
		model == QWEN_TURBO_LATEST
}
