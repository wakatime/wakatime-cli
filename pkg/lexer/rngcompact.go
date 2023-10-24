package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RNGCompact lexer. For RelaxNG-compact <http://relaxng.org> syntax.
type RNGCompact struct{}

// Lexer returns the lexer.
func (l RNGCompact) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rnc", "rng-compact"},
			Filenames: []string{"*.rnc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RNGCompact) Name() string {
	return heartbeat.LanguageRNGCompact.StringChroma()
}
