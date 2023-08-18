package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Nemerle lexer.
type Nemerle struct{}

// Lexer returns the lexer.
func (l Nemerle) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"nemerle"},
			Filenames: []string{"*.n"},
			// inferred
			MimeTypes: []string{"text/x-nemerle"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// Nemerle is quite similar to Python, but @if is relatively uncommon
		// elsewhere.
		if strings.Contains(text, "@if") {
			return 0.1
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Nemerle) Name() string {
	return heartbeat.LanguageNemerle.StringChroma()
}
