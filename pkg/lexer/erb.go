package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ERB lexer.
type ERB struct{}

// Lexer returns the lexer.
func (l ERB) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"erb"},
			MimeTypes: []string{"application/x-ruby-templating"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if strings.Contains(text, "<%") && strings.Contains(text, "%>") {
			return 0.4
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (ERB) Name() string {
	return heartbeat.LanguageERB.StringChroma()
}
