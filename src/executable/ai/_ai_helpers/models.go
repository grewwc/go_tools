package _ai_helpers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/grewwc/go_tools/src/terminalW"
	"github.com/grewwc/go_tools/src/utilsW"
)

const (
	DEEPSEEK               = "deepseek-r1"
	QWEN_MAX_LASTEST       = "qwen-max-latest"
	QWEN_PLUS              = "qwen-plus"
	QWEN_MAX               = "qwen-max"
	QWEN_CODER_PLUS_LATEST = "qwen-coder-plus-latest"
	QWEN_LONG              = "qwen-long"
	QWQ                    = "qwq-plus-latest"
)

func getDefaultModel() string {
	return QWQ
}

func GetModel(parsed *terminalW.ParsedResults) string {
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

var NonTextFile = utilsW.NewThreadSafeVal([]string{})

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
		return GetModel(terminalW.ParseArgs(fmt.Sprintf("a %s", found)))
	}

	return prevModel
}

func SearchEnabled(model string) bool {
	return model == QWEN_MAX || model == QWEN_MAX_LASTEST || model == QWEN_PLUS
}
