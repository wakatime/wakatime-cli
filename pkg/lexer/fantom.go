package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Fantom lexer.
type Fantom struct{}

// Lexer returns the lexer.
func (l Fantom) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"fan"},
			Filenames: []string{"*.fan"},
			MimeTypes: []string{"application/x-fantom"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Fantom) Name() string {
	return heartbeat.LanguageFantom.StringChroma()
}
