package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ADL lexer.
type ADL struct{}

// Lexer returns the lexer.
func (l ADL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"adl"},
			Filenames: []string{"*.adl", "*.adls", "*.adlf", "*.adlx"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ADL) Name() string {
	return heartbeat.LanguageADL.StringChroma()
}
