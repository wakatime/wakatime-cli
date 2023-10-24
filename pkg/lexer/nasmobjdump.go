package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// NASMObjdump lexer.
type NASMObjdump struct{}

// Lexer returns the lexer.
func (l NASMObjdump) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"objdump-nasm"},
			Filenames: []string{"*.objdump-intel"},
			MimeTypes: []string{"text/x-nasm-objdump"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (NASMObjdump) Name() string {
	return heartbeat.LanguageNASMObjdump.StringChroma()
}
