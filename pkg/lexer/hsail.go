package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// HSAIL lexer.
type HSAIL struct{}

// Lexer returns the lexer.
func (l HSAIL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"hsail", "hsa"},
			Filenames: []string{"*.hsail"},
			MimeTypes: []string{"text/x-hsail"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (HSAIL) Name() string {
	return heartbeat.LanguageHSAIL.StringChroma()
}
