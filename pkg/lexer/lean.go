package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Lean lexer.
type Lean struct{}

// Lexer returns the lexer.
func (l Lean) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"lean"},
			Filenames: []string{"*.lean"},
			MimeTypes: []string{"text/x-lean"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Lean) Name() string {
	return heartbeat.LanguageLean.StringChroma()
}
