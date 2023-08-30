package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Snowball lexer. Lexer for Snowball <http://snowballstem.org/> source code.
type Snowball struct{}

// Lexer returns the lexer.
func (l Snowball) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"snowball"},
			Filenames: []string{"*.sbl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Snowball) Name() string {
	return heartbeat.LanguageSnowball.StringChroma()
}
