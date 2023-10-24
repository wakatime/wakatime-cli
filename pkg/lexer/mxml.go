package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MXML lexer.
type MXML struct{}

// Lexer returns the lexer.
func (l MXML) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mxml"},
			Filenames: []string{"*.mxml"},
			MimeTypes: []string{"text/xml", "application/xml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MXML) Name() string {
	return heartbeat.LanguageMXML.StringChroma()
}
