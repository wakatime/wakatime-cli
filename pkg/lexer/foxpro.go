package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// FoxPro lexer.
type FoxPro struct{}

// Lexer returns the lexer.
func (l FoxPro) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"foxpro", "vfp", "clipper", "xbase"},
			Filenames: []string{"*.PRG", "*.prg"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (FoxPro) Name() string {
	return heartbeat.LanguageFoxPro.StringChroma()
}
