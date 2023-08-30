package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Redcode lexer.
type Redcode struct{}

// Lexer returns the lexer.
func (l Redcode) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"redcode"},
			Filenames: []string{"*.cw"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Redcode) Name() string {
	return heartbeat.LanguageRedcode.StringChroma()
}
