package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Felix lexer.
type Felix struct{}

// Lexer returns the lexer.
func (l Felix) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"felix", "flx"},
			Filenames: []string{"*.flx", "*.flxh"},
			MimeTypes: []string{"text/x-felix"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Felix) Name() string {
	return heartbeat.LanguageFelix.StringChroma()
}
