package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Icon lexer.
type Icon struct{}

// Lexer returns the lexer.
func (l Icon) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"icon"},
			Filenames: []string{"*.icon", "*.ICON"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Icon) Name() string {
	return heartbeat.LanguageIcon.StringChroma()
}
