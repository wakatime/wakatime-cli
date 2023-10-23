package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// XQuery lexer.
type XQuery struct{}

// Lexer returns the lexer.
func (l XQuery) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"xquery", "xqy", "xq", "xql", "xqm"},
			Filenames: []string{"*.xqy", "*.xquery", "*.xq", "*.xql", "*.xqm"},
			MimeTypes: []string{"text/xquery", "application/xquery"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (XQuery) Name() string {
	return heartbeat.LanguageXQuery.StringChroma()
}
