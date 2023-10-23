package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/doctype"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// HTML lexer.
type HTML struct{}

// Lexer returns the lexer.
func (l HTML) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if matched, _ := doctype.MatchString(text, "html"); matched {
				return 0.5
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (HTML) Name() string {
	return heartbeat.LanguageHTML.StringChroma()
}
