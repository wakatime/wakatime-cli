package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// VGL lexer.
type VGL struct{}

// Lexer returns the lexer.
func (l VGL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"vgl"},
			Filenames: []string{"*.rpf"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (VGL) Name() string {
	return heartbeat.LanguageVGL.StringChroma()
}
