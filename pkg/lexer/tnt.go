package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// TNT lexer. Lexer for Typographic Number Theory, as described in the book
// Gödel, Escher, Bach, by Douglas R. Hofstadter, or as summarized here:
// https://github.com/Kenny2github/language-tnt/blob/master/README.md#summary-of-tnt
type TNT struct{}

// Lexer returns the lexer.
func (l TNT) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"tnt"},
			Filenames: []string{"*.tnt"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (TNT) Name() string {
	return heartbeat.LanguageTNT.StringChroma()
}
