package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CAmkES lexer.
type CAmkES struct{}

// Lexer returns the lexer.
func (l CAmkES) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"camkes", "idl4"},
			Filenames: []string{"*.camkes", "*.idl4"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CAmkES) Name() string {
	return heartbeat.LanguageCAmkES.StringChroma()
}
