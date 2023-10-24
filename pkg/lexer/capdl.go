package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CapDL lexer.
type CapDL struct{}

// Lexer returns the lexer.
func (l CapDL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"capdl"},
			Filenames: []string{"*.cdl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CapDL) Name() string {
	return heartbeat.LanguageCapDL.StringChroma()
}
