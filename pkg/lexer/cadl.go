package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CADL lexer.
type CADL struct{}

// Lexer returns the lexer.
func (l CADL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cadl"},
			Filenames: []string{"*.cadl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CADL) Name() string {
	return heartbeat.LanguageCADL.StringChroma()
}
