package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MQL lexer.
type MQL struct{}

// Lexer returns the lexer.
func (l MQL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"mql", "mq4", "mq5", "mql4", "mql5"},
			Filenames: []string{"*.mq4", "*.mq5", "*.mqh"},
			MimeTypes: []string{"text/x-mql"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MQL) Name() string {
	return heartbeat.LanguageMQL.StringChroma()
}
