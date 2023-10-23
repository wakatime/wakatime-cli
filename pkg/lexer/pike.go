package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Pike lexer.
type Pike struct{}

// Lexer returns the lexer.
func (l Pike) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"pike"},
			Filenames: []string{"*.pike", "*.pmod"},
			MimeTypes: []string{"text/x-pike"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Pike) Name() string {
	return heartbeat.LanguagePike.StringChroma()
}
