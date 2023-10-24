package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Golo lexer.
type Golo struct{}

// Lexer returns the lexer.
func (l Golo) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"golo"},
			Filenames: []string{"*.golo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Golo) Name() string {
	return heartbeat.LanguageGolo.StringChroma()
}
