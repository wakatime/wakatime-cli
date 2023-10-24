package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// DASM16 lexer.
type DASM16 struct{}

// Lexer returns the lexer.
func (l DASM16) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"dasm16"},
			Filenames: []string{"*.dasm16", "*.dasm"},
			MimeTypes: []string{"text/x-dasm16"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (DASM16) Name() string {
	return heartbeat.LanguageDASM16.StringChroma()
}
