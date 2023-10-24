package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Scdoc lexer.
type Scdoc struct{}

// Lexer returns the lexer.
func (l Scdoc) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"scdoc", "scd"},
			Filenames: []string{"*.scd", "*.scdoc"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// This is very similar to markdown, save for the escape characters
		// needed for * and _.
		var result float32

		if strings.Contains(text, `\*`) {
			result += 0.01
		}

		if strings.Contains(text, `\_`) {
			result += 0.01
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (Scdoc) Name() string {
	return heartbeat.LanguageScdoc.StringChroma()
}
