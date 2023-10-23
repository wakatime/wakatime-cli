package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// TcshSession lexer. Lexer for Tcsh sessions, i.e. command lines, including a
// prompt, interspersed with output.
type TcshSession struct{}

// Lexer returns the lexer.
func (l TcshSession) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"tcshcon"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (TcshSession) Name() string {
	return heartbeat.LanguageTcshSession.StringChroma()
}
