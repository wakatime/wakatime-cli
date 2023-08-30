package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// FloScript lexer.
type FloScript struct{}

// Lexer returns the lexer.
func (l FloScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"floscript", "flo"},
			Filenames: []string{"*.flo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (FloScript) Name() string {
	return heartbeat.LanguageFloScript.StringChroma()
}
