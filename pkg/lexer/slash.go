package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Slash lexer. Lexer for the Slash programming language.
type Slash struct{}

// Lexer returns the lexer.
func (l Slash) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"slash"},
			Filenames: []string{"*.sla"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Slash) Name() string {
	return heartbeat.LanguageSlash.StringChroma()
}
