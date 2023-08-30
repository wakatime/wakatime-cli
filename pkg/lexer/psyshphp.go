package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// PsyShPHP lexer.
type PsyShPHP struct{}

// Lexer returns the lexer.
func (l PsyShPHP) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"psysh"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (PsyShPHP) Name() string {
	return heartbeat.LanguagePsyShPHP.StringChroma()
}
