package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CObjdump lexer.
type CObjdump struct{}

// Lexer returns the lexer.
func (l CObjdump) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"c-objdump"},
			Filenames: []string{"*.c-objdump"},
			MimeTypes: []string{"text/x-c-objdump"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CObjdump) Name() string {
	return heartbeat.LanguageCObjdump.StringChroma()
}
