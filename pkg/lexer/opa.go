package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Opa lexer.
type Opa struct{}

// Lexer returns the lexer.
func (l Opa) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"opa"},
			Filenames: []string{"*.opa"},
			MimeTypes: []string{"text/x-opa"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Opa) Name() string {
	return heartbeat.LanguageOpa.StringChroma()
}
