package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MSDOSSession lexer.
type MSDOSSession struct{}

// Lexer returns the lexer.
func (l MSDOSSession) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"doscon"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MSDOSSession) Name() string {
	return heartbeat.LanguageMSDOSSession.StringChroma()
}
