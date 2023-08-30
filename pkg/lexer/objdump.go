package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Objdump lexer.
type Objdump struct{}

// Lexer returns the lexer.
func (l Objdump) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"objdump"},
			Filenames: []string{"*.objdump"},
			MimeTypes: []string{"text/x-objdump"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Objdump) Name() string {
	return heartbeat.LanguageObjdump.StringChroma()
}
