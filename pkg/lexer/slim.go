package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Slim lexer.
type Slim struct{}

// Lexer returns the lexer.
func (l Slim) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"slim"},
			Filenames: []string{"*.slim"},
			MimeTypes: []string{"text/x-slim"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Slim) Name() string {
	return heartbeat.LanguageSlim.StringChroma()
}
