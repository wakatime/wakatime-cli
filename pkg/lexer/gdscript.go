package lexer

import (
	"math"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	gdscriptAnalyserFuncRe     = regexp.MustCompile(`func (_ready|_init|_input|_process|_unhandled_input)`)
	gdscriptAnalyserKeywordRe  = regexp.MustCompile(`(extends |class_name |onready |preload|load|setget|func [^_])`)
	gdscriptAnalyserKeyword2Re = regexp.MustCompile(`(var|const|enum|export|signal|tool)`)
)

// GDScript lexer.
type GDScript struct{}

// Lexer returns the lexer.
func (l GDScript) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
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

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (GDScript) Name() string {
	return heartbeat.LanguageGDScript.StringChroma()
}
