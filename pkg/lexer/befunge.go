package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Befunge lexer.
type Befunge struct{}

// Lexer returns the lexer.
func (l Befunge) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"befunge"},
			Filenames: []string{"*.befunge"},
			MimeTypes: []string{"application/x-befunge"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Befunge) Name() string {
	return heartbeat.LanguageBefunge.StringChroma()
}
