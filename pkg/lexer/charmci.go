package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Charmci lexer.
type Charmci struct{}

// Lexer returns the lexer.
func (l Charmci) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"charmci"},
			Filenames: []string{"*.ci"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Charmci) Name() string {
	return heartbeat.LanguageCharmci.StringChroma()
}
