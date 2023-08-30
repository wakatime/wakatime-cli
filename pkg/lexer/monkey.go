package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Monkey lexer.
type Monkey struct{}

// Lexer returns the lexer.
func (l Monkey) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"monkey"},
			Filenames: []string{"*.monkey"},
			MimeTypes: []string{"text/x-monkey"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Monkey) Name() string {
	return heartbeat.LanguageMonkey.StringChroma()
}
