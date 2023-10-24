package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LiterateCryptol lexer.
type LiterateCryptol struct{}

// Lexer returns the lexer.
func (l LiterateCryptol) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"lcry", "literate-cryptol", "lcryptol"},
			Filenames: []string{"*.lcry"},
			MimeTypes: []string{"text/x-literate-cryptol"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LiterateCryptol) Name() string {
	return heartbeat.LanguageLiterateCryptol.StringChroma()
}
