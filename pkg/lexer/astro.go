package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Astro lexer.
type Astro struct{}

// Lexer returns the lexer.
func (l Astro) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"astro"},
			Filenames: []string{"*.astro"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Astro) Name() string {
	return heartbeat.LanguageAstro.StringChroma()
}
