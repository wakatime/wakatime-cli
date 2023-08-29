package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Inform6Template lexer.
type Inform6Template struct{}

// Lexer returns the lexer.
func (l Inform6Template) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"i6t"},
			Filenames: []string{"*.i6t"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Inform6Template) Name() string {
	return heartbeat.LanguageInform6Template.StringChroma()
}
