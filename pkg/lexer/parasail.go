package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ParaSail lexer.
type ParaSail struct{}

// Lexer returns the lexer.
func (l ParaSail) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"parasail"},
			Filenames: []string{"*.psi", "*.psl"},
			MimeTypes: []string{"text/x-parasail"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ParaSail) Name() string {
	return heartbeat.LanguageParaSail.StringChroma()
}
