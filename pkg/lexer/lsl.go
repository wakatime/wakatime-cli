package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LSL lexer.
type LSL struct{}

// Lexer returns the lexer.
func (l LSL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"lsl"},
			Filenames: []string{"*.lsl"},
			MimeTypes: []string{"text/x-lsl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LSL) Name() string {
	return heartbeat.LanguageLSL.StringChroma()
}
