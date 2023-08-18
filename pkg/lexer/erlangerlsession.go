package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ErlangErlSession lexer.
type ErlangErlSession struct{}

// Lexer returns the lexer.
func (l ErlangErlSession) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"erl"},
			Filenames: []string{"*.erl-sh"},
			MimeTypes: []string{"text/x-erl-shellsession"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ErlangErlSession) Name() string {
	return heartbeat.LanguageErlangErlSession.StringChroma()
}
