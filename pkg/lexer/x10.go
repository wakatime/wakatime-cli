package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// X10 lexer.
type X10 struct{}

// Lexer returns the lexer.
func (l X10) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"x10", "xten"},
			Filenames: []string{"*.x10"},
			MimeTypes: []string{"text/x-x10"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (X10) Name() string {
	return heartbeat.LanguageX10.StringChroma()
}
