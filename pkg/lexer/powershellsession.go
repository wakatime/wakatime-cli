package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// PowerShellSession lexer.
type PowerShellSession struct{}

// Lexer returns the lexer.
func (l PowerShellSession) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"ps1con"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (PowerShellSession) Name() string {
	return heartbeat.LanguagePowerShellSession.StringChroma()
}
