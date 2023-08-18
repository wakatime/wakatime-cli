package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Freefem lexer.
type Freefem struct{}

// Lexer returns the lexer.
func (l Freefem) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"freefem"},
			Filenames: []string{"*.edp"},
			MimeTypes: []string{"text/x-freefem"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Freefem) Name() string {
	return heartbeat.LanguageFreefem.StringChroma()
}
