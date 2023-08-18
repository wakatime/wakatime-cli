package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// DylanSession lexer.
type DylanSession struct{}

// Lexer returns the lexer.
func (l DylanSession) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"dylan-console", "dylan-repl"},
			Filenames: []string{"*.dylan-console"},
			MimeTypes: []string{"text/x-dylan-console"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (DylanSession) Name() string {
	return heartbeat.LanguageDylanSession.StringChroma()
}
