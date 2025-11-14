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
	DEEPSEEK_V3            = "deepseek-v3.1"
	DEEPSEEK_R1            = "deepseek-r1"
	QWEN_MAX_LASTEST       = "qwen-max-latest"
	QWEN_PLUS_LATEST       = "qwen-plus-latest"
	QWEN_MAX               = "qwen-max"
	QWEN_CODER_PLUS_LATEST = "qwen3-coder-plus"
	QWEN_LONG              = "qwen-long"
	QWQ                    = "qwq-plus-latest"
	QWEN_FLASH             = "qwen-flash"
	QWEN3_MAX              = "qwen3-max"
)

const (
	QWEN_VL_FLASH = "qwen3-vl-flash"
	QWEN_VL_MAX   = "qwen3-vl-plus"
	QWEN_VL_OCR   = "qwen-vl-ocr-latest"
)

var allModels = cw.NewSetT(
	DEEPSEEK_V3, QWEN_MAX_LASTEST, QWEN_PLUS_LATEST, QWEN_MAX,
	QWEN_CODER_PLUS_LATEST, QWEN_LONG, QWQ, QWEN_FLASH, QWEN3_MAX,
)

var allVlModels = cw.NewSetT(
	QWEN_VL_FLASH, QWEN_VL_MAX, QWEN_VL_OCR,
)

var enableSearchModels = cw.NewSetT(
	QWEN_MAX, QWEN_MAX_LASTEST, QWEN_PLUS_LATEST, QWEN_FLASH, QWEN_PLUS_LATEST,
	DEEPSEEK_V3, QWEN3_MAX,
)

func init() {
	allModels = allModels.Union(allVlModels)
}

func isVlModel(model string) bool {
	return allVlModels.Contains(model)
}

const (
	QWEN_ENDPOINT = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
)

func getDefaultModel() string {
	config := utilsw.GetAllConfig()
	return config.GetOrDefault("ai.model.default", QWEN3_MAX).(string)
}

func getEndpoint() string {
	config := utilsw.GetAllConfig()
	return config.GetOrDefault("ai.model.endpoint", QWEN_ENDPOINT).(string)
}

func GetModel(parsed *terminalw.Parser) string {
	if parsed.ContainsFlagStrict("code") {
		return QWEN_CODER_PLUS_LATEST
	}

	thinkingMode := parsed.ContainsFlagStrict("t")
	if parsed.ContainsFlagStrict("d") {
		if thinkingMode {
			return DEEPSEEK_R1
		}
		return DEEPSEEK_V3
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
		if thinkingMode {
			return DEEPSEEK_R1
		}
		return DEEPSEEK_V3
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
	case DEEPSEEK_V3, "5":
		if thinkingMode {
			return DEEPSEEK_R1
		}
		return DEEPSEEK_V3
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
	if len(NonTextFile.Get().([]string)) > 0 && !isVlModel(prevModel) {
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
		return DEEPSEEK_V3
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

func searchEnabled(model string) bool {
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

