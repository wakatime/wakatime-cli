package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var logosAnalyserKeywordsRe = regexp.MustCompile(`%(?:hook|ctor|init|c\()`)

// Logos lexer.
type Logos struct{}

// Lexer returns the lexer.
func (l Logos) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"logos"},
			Filenames: []string{"*.x", "*.xi", "*.xm", "*.xmi"},
			MimeTypes: []string{"text/x-logos"},
			Priority:  0.25,
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if logosAnalyserKeywordsRe.MatchString(text) {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Logos) Name() string {
	return heartbeat.LanguageLogos.StringChroma()
}
