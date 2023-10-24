package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Praat lexer.
type Praat struct{}

// Lexer returns the lexer.
func (l Praat) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"praat"},
			Filenames: []string{"*.praat", "*.proc", "*.psc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Praat) Name() string {
	return heartbeat.LanguagePraat.StringChroma()
}
