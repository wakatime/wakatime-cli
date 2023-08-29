package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Xtend lexer.
type Xtend struct{}

// Lexer returns the lexer.
func (l Xtend) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"xtend"},
			Filenames: []string{"*.xtend"},
			MimeTypes: []string{"text/x-xtend"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Xtend) Name() string {
	return heartbeat.LanguageXtend.StringChroma()
}
