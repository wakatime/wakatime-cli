package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Isabelle lexer.
type Isabelle struct{}

// Lexer returns the lexer.
func (l Isabelle) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"isabelle"},
			Filenames: []string{"*.thy"},
			MimeTypes: []string{"text/x-isabelle"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Isabelle) Name() string {
	return heartbeat.LanguageIsabelle.StringChroma()
}
