package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Ooc lexer.
type Ooc struct{}

// Lexer returns the lexer.
func (l Ooc) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ooc"},
			Filenames: []string{"*.ooc"},
			MimeTypes: []string{"text/x-ooc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Ooc) Name() string {
	return heartbeat.LanguageOoc.StringChroma()
}
