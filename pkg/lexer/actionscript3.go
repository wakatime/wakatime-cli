package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoglobals
var actionscript3AnalyserRe = regexp.MustCompile(`\w+\s*:\s*\w`)

// ActionScript3 lexer.
type ActionScript3 struct{}

// Lexer returns the lexer.
func (l ActionScript3) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if actionscript3AnalyserRe.MatchString(text) {
				return 0.3
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (ActionScript3) Name() string {
	return heartbeat.LanguageActionScript.StringChroma()
}
