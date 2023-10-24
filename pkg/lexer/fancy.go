package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Fancy lexer.
type Fancy struct{}

// Lexer returns the lexer.
func (l Fancy) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"fancy", "fy"},
			Filenames: []string{"*.fy", "*.fancypack"},
			MimeTypes: []string{"text/x-fancysrc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Fancy) Name() string {
	return heartbeat.LanguageFancy.StringChroma()
}
