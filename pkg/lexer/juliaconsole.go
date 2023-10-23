package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// JuliaConsole lexer.
type JuliaConsole struct{}

// Lexer returns the lexer.
func (l JuliaConsole) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"jlcon"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (JuliaConsole) Name() string {
	return heartbeat.LanguageJuliaConsole.StringChroma()
}
