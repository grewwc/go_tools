package internal

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/grewwc/go_tools/src/algow"
	"github.com/grewwc/go_tools/src/cw"
	"github.com/grewwc/go_tools/src/strw"
	"github.com/grewwc/go_tools/src/terminalw"
	"github.com/grewwc/go_tools/src/utilsw"
)

const (
	DEEPSEEK               = "deepseek-r1-0528"
	QWEN_MAX_LASTEST       = "qwen-max-latest"
	QWEN_PLUS_LATEST       = "qwen-plus-latest"
	QWEN_MAX               = "qwen-max"
	QWEN_CODER_PLUS_LATEST = "qwen3-coder-plus"
	QWEN_LONG              = "qwen-long"
	QWQ                    = "qwq-plus-latest"
	QWEN_FLASH             = "qwen-flash"
	QWEN3_MAX              = "qwen3-max"
)

var allModels = cw.NewSetT(
	DEEPSEEK, QWEN_MAX_LASTEST, QWEN_PLUS_LATEST, QWEN_MAX,
	QWEN_CODER_PLUS_LATEST, QWEN_LONG, QWQ, QWEN_FLASH, QWEN3_MAX,
)

const (
	QWEN_ENDPOINT = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
)

func getDefaultModel() string {
	config := utilsw.GetAllConfig()
	return config.GetOrDefault("ai.model.default", QWEN3_MAX).(string)
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
		return QWEN_PLUS_LATEST
	case 2:
		return QWEN_MAX
	case 3:
		return QWEN3_MAX
	case 4:
		return QWEN_CODER_PLUS_LATEST
	case 5:
		return DEEPSEEK
	case 6:
		return QWEN_FLASH
	}
	model := parsed.GetFlagValueDefault("m", getDefaultModel())

	switch model {
	case QWQ, "0":
		return QWQ
	case QWEN_PLUS_LATEST, "1":
		return QWEN_PLUS_LATEST
	case QWEN_MAX, "2":
		return QWEN_MAX
	case QWEN3_MAX, "3":
		return QWEN3_MAX
	case QWEN_CODER_PLUS_LATEST, "4":
		return QWEN_CODER_PLUS_LATEST
	case DEEPSEEK, "5":
		return DEEPSEEK
	case QWEN_FLASH, "6":
		return QWEN_FLASH
	default:
		if !strw.IsBlank(model) {
			return determinModel(model)
		}
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

var enableSearchModels = cw.NewSet(
	QWEN_MAX, QWEN_MAX_LASTEST, QWEN_PLUS_LATEST, QWEN_FLASH, QWEN_PLUS_LATEST,
	DEEPSEEK, QWEN3_MAX,
)

func SearchEnabled(model string) bool {
	return enableSearchModels.Contains(model)
}

func determinModel(model string) string {
	model = strings.ToLower(model)
	result := model
	dist := float32(math.MaxFloat32)
	for m := range allModels.Iter().Iterate() {
		currDist := float32(algow.EditDistance([]byte(m), []byte(model), nil)) / float32(len(model)+len(m))
		if currDist < dist {
			dist = currDist
			result = m
		}
	}
	return result
}
