package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Cypher lexer.
type Cypher struct{}

// Lexer returns the lexer.
func (l Cypher) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cypher"},
			Filenames: []string{"*.cyp", "*.cypher"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Cypher) Name() string {
	return heartbeat.LanguageCypher.StringChroma()
}
