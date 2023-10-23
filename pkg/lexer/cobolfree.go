package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// COBOLFree lexer.
type COBOLFree struct{}

// Lexer returns the lexer.
func (l COBOLFree) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cobolfree"},
			Filenames: []string{"*.cbl", "*.CBL"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (COBOLFree) Name() string {
	return heartbeat.LanguageCOBOLFree.StringChroma()
}
