package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Xtlang lexer.
type Xtlang struct{}

// Lexer returns the lexer.
func (l Xtlang) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"extempore"},
			Filenames: []string{"*.xtm"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Xtlang) Name() string {
	return heartbeat.LanguageXtlang.StringChroma()
}
