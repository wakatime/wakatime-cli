package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Kal lexer.
type Kal struct{}

// Lexer returns the lexer.
func (l Kal) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"kal"},
			Filenames: []string{"*.kal"},
			MimeTypes: []string{"text/kal", "application/kal"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Kal) Name() string {
	return heartbeat.LanguageKal.StringChroma()
}
