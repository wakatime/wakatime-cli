package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Cirru lexer.
type Cirru struct{}

// Lexer returns the lexer.
func (l Cirru) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cirru"},
			Filenames: []string{"*.cirru"},
			MimeTypes: []string{"text/x-cirru"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Cirru) Name() string {
	return heartbeat.LanguageCirru.StringChroma()
}
