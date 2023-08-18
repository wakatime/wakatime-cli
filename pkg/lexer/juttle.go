package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Juttle lexer.
type Juttle struct{}

// Lexer returns the lexer.
func (l Juttle) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"juttle"},
			Filenames: []string{"*.juttle"},
			MimeTypes: []string{"application/juttle", "application/x-juttle", "text/x-juttle", "text/juttle"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Juttle) Name() string {
	return heartbeat.LanguageJuttle.StringChroma()
}
