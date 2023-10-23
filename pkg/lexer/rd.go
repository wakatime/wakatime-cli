package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Rd lexer. Lexer for R documentation (Rd) files.
type Rd struct{}

// Lexer returns the lexer.
func (l Rd) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rd"},
			Filenames: []string{"*.Rd"},
			MimeTypes: []string{"text/x-r-doc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Rd) Name() string {
	return heartbeat.LanguageRd.StringChroma()
}
