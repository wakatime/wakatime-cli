package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Asymptote lexer.
type Asymptote struct{}

// Lexer returns the lexer.
func (l Asymptote) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"asy", "asymptote"},
			Filenames: []string{"*.asy"},
			MimeTypes: []string{"text/x-asymptote"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Asymptote) Name() string {
	return heartbeat.LanguageAsymptote.StringChroma()
}
