package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// NuSMV lexer.
type NuSMV struct{}

// Lexer returns the lexer.
func (l NuSMV) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"nusmv"},
			Filenames: []string{"*.smv"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (NuSMV) Name() string {
	return heartbeat.LanguageNuSMV.StringChroma()
}
