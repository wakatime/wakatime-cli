package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Flatline lexer.
type Flatline struct{}

// Lexer returns the lexer.
func (l Flatline) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"flatline"},
			MimeTypes: []string{"text/x-flatline"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Flatline) Name() string {
	return heartbeat.LanguageFlatline.StringChroma()
}
