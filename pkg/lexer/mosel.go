package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Mosel lexer.
type Mosel struct{}

// Lexer returns the lexer.
func (l Mosel) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"model"},
			Filenames: []string{"*.mos"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Mosel) Name() string {
	return heartbeat.LanguageMosel.StringChroma()
}
