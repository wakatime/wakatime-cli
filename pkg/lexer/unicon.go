package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Unicon lexer.
type Unicon struct{}

// Lexer returns the lexer.
func (l Unicon) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"unicon"},
			Filenames: []string{"*.icn"},
			MimeTypes: []string{"text/unicon"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Unicon) Name() string {
	return heartbeat.LanguageUnicon.StringChroma()
}
