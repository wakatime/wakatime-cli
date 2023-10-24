package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// WDiff lexer.
type WDiff struct{}

// Lexer returns the lexer.
func (l WDiff) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"wdiff"},
			Filenames: []string{"*.wdiff"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (WDiff) Name() string {
	return heartbeat.LanguageWDiff.StringChroma()
}
