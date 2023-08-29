package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ECL lexer.
type ECL struct{}

// Lexer returns the lexer.
func (l ECL) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ecl"},
			Filenames: []string{"*.ecl"},
			MimeTypes: []string{"application/x-ecl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// This is very difficult to guess relative to other business languages.
		// -> in conjunction with BEGIN/END seems relatively rare though.

		var result float32

		if strings.Contains(text, "->") {
			result += 0.01
		}

		if strings.Contains(text, "BEGIN") {
			result += 0.01
		}

		if strings.Contains(text, "END") {
			result += 0.01
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (ECL) Name() string {
	return heartbeat.LanguageECL.StringChroma()
}
