package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ColdfusionHTML lexer.
type ColdfusionHTML struct{}

// Lexer returns the lexer.
func (l ColdfusionHTML) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cfm"},
			Filenames: []string{"*.cfm", "*.cfml"},
			MimeTypes: []string{"application/x-coldfusion"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ColdfusionHTML) Name() string {
	return heartbeat.LanguageColdfusionHTML.StringChroma()
}
