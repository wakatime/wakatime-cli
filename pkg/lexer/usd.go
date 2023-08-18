package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// USD lexer.
type USD struct{}

// Lexer returns the lexer.
func (l USD) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"usd", "usda"},
			Filenames: []string{"*.usd", "*.usda"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (USD) Name() string {
	return heartbeat.LanguageUSD.StringChroma()
}
