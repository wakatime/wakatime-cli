package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// JSGF lexer.
type JSGF struct{}

// Lexer returns the lexer.
func (l JSGF) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"jsgf"},
			Filenames: []string{"*.jsgf"},
			MimeTypes: []string{"application/jsgf", "application/x-jsgf", "text/jsgf"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (JSGF) Name() string {
	return heartbeat.LanguageJSGF.StringChroma()
}
