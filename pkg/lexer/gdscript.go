package lexer

import (
	"math"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	gdscriptAnalyserFuncRe     = regexp.MustCompile(`func (_ready|_init|_input|_process|_unhandled_input)`)
	gdscriptAnalyserKeywordRe  = regexp.MustCompile(`(extends |class_name |onready |preload|load|setget|func [^_])`)
	gdscriptAnalyserKeyword2Re = regexp.MustCompile(`(var|const|enum|export|signal|tool)`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageGDScript.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		var result float64

		if gdscriptAnalyserFuncRe.MatchString(text) {
			result += 0.8
		}

		if gdscriptAnalyserKeywordRe.MatchString(text) {
			result += 0.4
		}

		if gdscriptAnalyserKeyword2Re.MatchString(text) {
			result += 0.2
		}

		return float32(math.Min(result, float64(1.0)))
	})
}
