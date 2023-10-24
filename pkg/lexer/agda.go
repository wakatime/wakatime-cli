package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Agda lexer.
type Agda struct{}

// Lexer returns the lexer.
func (l Agda) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"agda"},
			Filenames: []string{"*.agda"},
			MimeTypes: []string{"text/x-agda"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Agda) Name() string {
	return heartbeat.LanguageAgda.StringChroma()
}
