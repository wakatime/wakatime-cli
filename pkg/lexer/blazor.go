package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Blazor lexer.
type Blazor struct{}

// Lexer returns the lexer.
func (l Blazor) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"blazor"},
			Filenames: []string{"*.razor"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Blazor) Name() string {
	return heartbeat.LanguageBlazor.StringChroma()
}
