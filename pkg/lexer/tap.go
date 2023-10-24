package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// TAP lexer. For Test Anything Protocol (TAP) output.
type TAP struct{}

// Lexer returns the lexer.
func (l TAP) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"tap"},
			Filenames: []string{"*.tap"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (TAP) Name() string {
	return heartbeat.LanguageTAP.StringChroma()
}
