package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CPSA lexer.
type CPSA struct{}

// Lexer returns the lexer.
func (l CPSA) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cpsa"},
			Filenames: []string{"*.cpsa"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CPSA) Name() string {
	return heartbeat.LanguageCPSA.StringChroma()
}
