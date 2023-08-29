package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// BlitzMax lexer.
type BlitzMax struct{}

// Lexer returns the lexer.
func (l BlitzMax) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"blitzmax", "bmax"},
			Filenames: []string{"*.bmx"},
			MimeTypes: []string{"text/x-bmx"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (BlitzMax) Name() string {
	return heartbeat.LanguageBlitzMax.StringChroma()
}
