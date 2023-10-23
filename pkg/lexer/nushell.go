package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Nushell lexer.
type Nushell struct{}

// Lexer returns the lexer.
func (l Nushell) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"Nu"},
			Filenames: []string{"*.nu"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Nushell) Name() string {
	return heartbeat.LanguageNushell.StringChroma()
}
