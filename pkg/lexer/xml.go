package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/xml"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// XML lexer.
type XML struct{}

// Lexer returns the lexer.
func (l XML) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if xml.MatchString(text) {
				return 0.45 // less than HTML.
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (XML) Name() string {
	return heartbeat.LanguageXML.StringChroma()
}
