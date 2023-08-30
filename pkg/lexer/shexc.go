package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ShExC lexer. Lexer for ShExC <https://shex.io/shex-semantics/#shexc> shape expressions language syntax.
type ShExC struct{}

// Lexer returns the lexer.
func (l ShExC) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"shexc", "shex"},
			Filenames: []string{"*.shex"},
			MimeTypes: []string{"text/shex"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ShExC) Name() string {
	return heartbeat.LanguageShExC.StringChroma()
}
