package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	gasAnalyzerDirectiveRe      = regexp.MustCompile(`(?m)^\.(text|data|section)`)
	gasAnalyzerOtherDirectiveRe = regexp.MustCompile(`(?m)^\.\w+`)
)

// Gas lexer.
type Gas struct{}

// Lexer returns the lexer.
func (l Gas) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if gasAnalyzerDirectiveRe.MatchString(text) {
				return 1.0
			}

			if gasAnalyzerOtherDirectiveRe.MatchString(text) {
				return 0.1
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Gas) Name() string {
	return heartbeat.LanguageGas.StringChroma()
}
