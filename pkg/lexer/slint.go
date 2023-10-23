package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Slint lexer. Lexer for the Slint programming language.
type Slint struct{}

// Lexer returns the lexer.
func (l Slint) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"slint"},
			Filenames: []string{"*.slint"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Slint) Name() string {
	return heartbeat.LanguageSlint.StringChroma()
}
