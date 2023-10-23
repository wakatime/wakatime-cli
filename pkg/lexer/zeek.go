package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Zeek lexer.
type Zeek struct{}

// Lexer returns the lexer.
func (l Zeek) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"zeek", "bro"},
			Filenames: []string{"*.zeek", "*.bro"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Zeek) Name() string {
	return heartbeat.LanguageZeek.StringChroma()
}
