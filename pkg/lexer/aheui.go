package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Aheui lexer.
type Aheui struct{}

// Lexer returns the lexer.
func (l Aheui) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"aheui"},
			Filenames: []string{"*.aheui"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Aheui) Name() string {
	return heartbeat.LanguageAheui.StringChroma()
}
