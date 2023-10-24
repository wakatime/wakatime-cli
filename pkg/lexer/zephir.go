package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Zephir lexer.
type Zephir struct{}

// Lexer returns the lexer.
func (l Zephir) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"zephir"},
			Filenames: []string{"*.zep"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Zephir) Name() string {
	return heartbeat.LanguageZephir.StringChroma()
}
