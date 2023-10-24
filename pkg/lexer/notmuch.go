package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Notmuch lexer.
type Notmuch struct{}

// Lexer returns the lexer.
func (l Notmuch) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"notmuch"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if strings.HasPrefix(text, "\fmessage{") {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Notmuch) Name() string {
	return heartbeat.LanguageNotmuch.StringChroma()
}
