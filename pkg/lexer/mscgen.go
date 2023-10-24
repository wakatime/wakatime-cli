package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Mscgen lexer.
type Mscgen struct{}

// Lexer returns the lexer.
func (l Mscgen) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mscgen", "msc"},
			Filenames: []string{"*.msc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Mscgen) Name() string {
	return heartbeat.LanguageMscgen.StringChroma()
}
