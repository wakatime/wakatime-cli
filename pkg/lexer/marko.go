package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Marko lexer.
type Marko struct{}

// Lexer returns the lexer.
func (l Marko) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"marko"},
			Filenames: []string{"*.marko"},
			MimeTypes: []string{"text/x-marko"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Marko) Name() string {
	return heartbeat.LanguageMarko.StringChroma()
}
