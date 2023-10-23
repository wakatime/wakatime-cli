package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MOOCode lexer.
type MOOCode struct{}

// Lexer returns the lexer.
func (l MOOCode) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"moocode", "moo"},
			Filenames: []string{"*.moo"},
			MimeTypes: []string{"text/x-moocode"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MOOCode) Name() string {
	return heartbeat.LanguageMOOCode.StringChroma()
}
