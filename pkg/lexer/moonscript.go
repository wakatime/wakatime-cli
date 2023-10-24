package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MoonScript lexer.
type MoonScript struct{}

// Lexer returns the lexer.
func (l MoonScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"moon", "moonscript"},
			MimeTypes: []string{"text/x-moonscript", "application/x-moonscript"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MoonScript) Name() string {
	return heartbeat.LanguageMoonScript.StringChroma()
}
