package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// DG lexer.
type DG struct{}

// Lexer returns the lexer.
func (l DG) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"dg"},
			Filenames: []string{"*.dg"},
			MimeTypes: []string{"text/x-dg"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (DG) Name() string {
	return heartbeat.LanguageDG.StringChroma()
}
