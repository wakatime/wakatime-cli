package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ColdfusionCFC lexer.
type ColdfusionCFC struct{}

// Lexer returns the lexer.
func (l ColdfusionCFC) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cfc"},
			Filenames: []string{"*.cfc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ColdfusionCFC) Name() string {
	return heartbeat.LanguageColdfusionCFC.StringChroma()
}
