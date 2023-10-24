package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Boa lexer.
type Boa struct{}

// Lexer returns the lexer.
func (l Boa) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"boa"},
			Filenames: []string{"*.boa"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Boa) Name() string {
	return heartbeat.LanguageBoa.StringChroma()
}
