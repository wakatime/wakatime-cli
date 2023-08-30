package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Nit lexer.
type Nit struct{}

// Lexer returns the lexer.
func (l Nit) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"nit"},
			Filenames: []string{"*.nit"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Nit) Name() string {
	return heartbeat.LanguageNit.StringChroma()
}
