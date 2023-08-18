package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Augeas lexer.
type Augeas struct{}

// Lexer returns the lexer.
func (l Augeas) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"augeas"},
			Filenames: []string{"*.aug"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Augeas) Name() string {
	return heartbeat.LanguageAugeas.StringChroma()
}
