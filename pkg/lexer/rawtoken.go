package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RawToken lexer.
type RawToken struct{}

// Lexer returns the lexer.
func (l RawToken) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"raw"},
			MimeTypes: []string{"application/x-pygments-tokens"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RawToken) Name() string {
	return heartbeat.LanguageRawToken.StringChroma()
}
