package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Mustache lexer.
type Mustache struct{}

// Lexer returns the lexer.
func (l Mustache) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mustache"},
			Filenames: []string{"*.mustache"},
			MimeTypes: []string{"text/x-mustache-template"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Mustache) Name() string {
	return heartbeat.LanguageMustache.StringChroma()
}
