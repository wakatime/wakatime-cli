package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Boo lexer.
type Boo struct{}

// Lexer returns the lexer.
func (l Boo) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"boo"},
			Filenames: []string{"*.boo"},
			MimeTypes: []string{"text/x-boo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Boo) Name() string {
	return heartbeat.LanguageBoo.StringChroma()
}
