package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// AMPL lexer.
type AMPL struct{}

// Lexer returns the lexer.
func (l AMPL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ampl"},
			Filenames: []string{"*.run"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (AMPL) Name() string {
	return heartbeat.LanguageAMPL.StringChroma()
}
