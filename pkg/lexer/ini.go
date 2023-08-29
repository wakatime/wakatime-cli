package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// INI lexer.
type INI struct{}

// Lexer returns the lexer.
func (l INI) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			npos := strings.Count(text, "\n")
			if npos < 3 {
				return 0
			}

			if text[0] == '[' && text[npos-1] == ']' {
				return 1
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (INI) Name() string {
	return heartbeat.LanguageINI.StringChroma()
}
