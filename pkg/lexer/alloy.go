package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Alloy lexer.
type Alloy struct{}

// Lexer returns the lexer.
func (l Alloy) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"alloy"},
			Filenames: []string{"*.als"},
			MimeTypes: []string{"text/x-alloy"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Alloy) Name() string {
	return heartbeat.LanguageAlloy.StringChroma()
}
