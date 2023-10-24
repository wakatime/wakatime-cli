package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Mojo lexer.
type Mojo struct{}

// Lexer returns the lexer.
func (l Mojo) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mojo"},
			Filenames: []string{"*.🔥", "*.mojo"},
			MimeTypes: []string{"text/x-mojo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Mojo) Name() string {
	return heartbeat.LanguageMojo.StringChroma()
}
