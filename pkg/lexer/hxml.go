package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Hxml lexer.
type Hxml struct{}

// Lexer returns the lexer.
func (l Hxml) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"haxeml", "hxml"},
			Filenames: []string{"*.hxml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Hxml) Name() string {
	return heartbeat.LanguageHxml.StringChroma()
}
