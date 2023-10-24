package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Haml lexer.
type Haml struct{}

// Lexer returns the lexer.
func (l Haml) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"haml"},
			Filenames: []string{"*.haml"},
			MimeTypes: []string{"text/x-haml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Haml) Name() string {
	return heartbeat.LanguageHaml.StringChroma()
}
