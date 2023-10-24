package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Pointless lexer.
type Pointless struct{}

// Lexer returns the lexer.
func (l Pointless) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"pointless"},
			Filenames: []string{"*.ptls"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Pointless) Name() string {
	return heartbeat.LanguagePointless.StringChroma()
}
