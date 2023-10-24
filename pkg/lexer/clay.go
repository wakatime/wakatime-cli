package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Clay lexer.
type Clay struct{}

// Lexer returns the lexer.
func (l Clay) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"clay"},
			Filenames: []string{"*.clay"},
			MimeTypes: []string{"text/x-clay"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Clay) Name() string {
	return heartbeat.LanguageClay.StringChroma()
}
