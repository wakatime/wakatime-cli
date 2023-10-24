package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// NSIS lexer.
type NSIS struct{}

// Lexer returns the lexer.
func (l NSIS) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"nsis", "nsi", "nsh"},
			Filenames: []string{"*.nsi", "*.nsh"},
			MimeTypes: []string{"text/x-nsis"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (NSIS) Name() string {
	return heartbeat.LanguageNSIS.StringChroma()
}
