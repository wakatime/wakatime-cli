package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Clean lexer.
type Clean struct{}

// Lexer returns the lexer.
func (l Clean) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"clean"},
			Filenames: []string{"*.icl", "*.dcl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Clean) Name() string {
	return heartbeat.LanguageClean.StringChroma()
}
