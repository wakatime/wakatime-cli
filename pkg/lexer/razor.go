package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Razor lexer. Lexer for Blazor's Razor files.
type Razor struct{}

// Lexer returns the lexer.
func (l Razor) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"razor"},
			Filenames: []string{"*.razor"},
			MimeTypes: []string{"text/html"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Razor) Name() string {
	return heartbeat.LanguageRazor.StringChroma()
}
