package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ReScript lexer.
type ReScript struct{}

// Lexer returns the lexer.
func (l ReScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rescript"},
			Filenames: []string{"*.res", "*.resi"},
			MimeTypes: []string{"text/x-rescript"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ReScript) Name() string {
	return heartbeat.LanguageReScript.StringChroma()
}
