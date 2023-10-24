package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Ioke lexer.
type Ioke struct{}

// Lexer returns the lexer.
func (l Ioke) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ioke", "ik"},
			Filenames: []string{"*.ik"},
			MimeTypes: []string{"text/x-iokesrc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Ioke) Name() string {
	return heartbeat.LanguageIoke.StringChroma()
}
