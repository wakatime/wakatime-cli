package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Inform7 lexer.
type Inform7 struct{}

// Lexer returns the lexer.
func (l Inform7) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"inform7", "i7"},
			Filenames: []string{"*.ni", "*.i7x"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Inform7) Name() string {
	return heartbeat.LanguageInform7.StringChroma()
}
