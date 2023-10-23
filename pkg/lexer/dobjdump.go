package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// DObjdump lexer.
type DObjdump struct{}

// Lexer returns the lexer.
func (l DObjdump) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"d-objdump"},
			Filenames: []string{"*.d-objdump"},
			MimeTypes: []string{"text/x-d-objdump"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (DObjdump) Name() string {
	return heartbeat.LanguageDObjdump.StringChroma()
}
