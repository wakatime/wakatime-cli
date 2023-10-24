package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MAQL lexer.
type MAQL struct{}

// Lexer returns the lexer.
func (l MAQL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"maql"},
			Filenames: []string{"*.maql"},
			MimeTypes: []string{"text/x-gooddata-maql", "application/x-gooddata-maql"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MAQL) Name() string {
	return heartbeat.LanguageMAQL.StringChroma()
}
