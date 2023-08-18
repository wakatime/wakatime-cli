package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// JSONLD lexer.
type JSONLD struct{}

// Lexer returns the lexer.
func (l JSONLD) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"jsonld", "json-ld"},
			Filenames: []string{"*.jsonld"},
			MimeTypes: []string{"application/ld+json"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (JSONLD) Name() string {
	return heartbeat.LanguageJSONLD.StringChroma()
}
