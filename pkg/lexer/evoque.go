package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Evoque lexer.
type Evoque struct{}

// Lexer returns the lexer.
func (l Evoque) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"evoque"},
			Filenames: []string{"*.evoque"},
			MimeTypes: []string{"application/x-evoque"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// Evoque templates use $evoque, which is unique.
		if strings.Contains(text, "$evoque") {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Evoque) Name() string {
	return heartbeat.LanguageEvoque.StringChroma()
}
