package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Mask lexer.
type Mask struct{}

// Lexer returns the lexer.
func (l Mask) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mask"},
			Filenames: []string{"*.mask"},
			MimeTypes: []string{"text/x-mask"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Mask) Name() string {
	return heartbeat.LanguageMask.StringChroma()
}
