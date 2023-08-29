package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MIME lexer.
type MIME struct{}

// Lexer returns the lexer.
func (l MIME) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mime"},
			MimeTypes: []string{"multipart/mixed", "multipart/related", "multipart/alternative"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MIME) Name() string {
	return heartbeat.LanguageMIME.StringChroma()
}
