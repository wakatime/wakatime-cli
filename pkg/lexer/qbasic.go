package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// QBasic lexer.
type QBasic struct{}

// Lexer returns the lexer.
func (l QBasic) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if strings.Contains(text, "$DYNAMIC") || strings.Contains(text, "$STATIC") {
				return 0.9
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (QBasic) Name() string {
	return heartbeat.LanguageQBasic.StringChroma()
}
