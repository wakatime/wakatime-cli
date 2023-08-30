package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var limboAnalyzerRe = regexp.MustCompile(`(?m)^implement \w+;`)

// Limbo lexer.
type Limbo struct{}

// Lexer returns the lexer.
func (l Limbo) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"limbo"},
			Filenames: []string{"*.b"},
			MimeTypes: []string{"text/limbo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// Any limbo module implements something
		if limboAnalyzerRe.MatchString(text) {
			return 0.7
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Limbo) Name() string {
	return heartbeat.LanguageLimbo.StringChroma()
}
