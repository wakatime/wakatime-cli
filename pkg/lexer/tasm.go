package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var tasmAnalyzerRe = regexp.MustCompile(`(?i)PROC`)

// TASM lexer.
type TASM struct{}

// Lexer returns the lexer.
func (l TASM) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if tasmAnalyzerRe.MatchString(text) {
				return 1.0
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (TASM) Name() string {
	return heartbeat.LanguageTASM.StringChroma()
}
