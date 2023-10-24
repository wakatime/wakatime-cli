package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MiniScript lexer.
type MiniScript struct{}

// Lexer returns the lexer.
func (l MiniScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ms", "miniscript"},
			Filenames: []string{"*.ms"},
			MimeTypes: []string{"text/x-miniscript", "application/x-miniscript"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MiniScript) Name() string {
	return heartbeat.LanguageMiniScript.StringChroma()
}
