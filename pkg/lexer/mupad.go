package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MuPAD lexer.
type MuPAD struct{}

// Lexer returns the lexer.
func (l MuPAD) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mupad"},
			Filenames: []string{"*.mu"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MuPAD) Name() string {
	return heartbeat.LanguageMuPAD.StringChroma()
}
