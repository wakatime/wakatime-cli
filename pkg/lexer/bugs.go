package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var bugsAnalyzerRe = regexp.MustCompile(`(?m)^\s*model\s*{`)

// BUGS lexer.
type BUGS struct{}

// Lexer returns the lexer.
func (l BUGS) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"bugs", "winbugs", "openbugs"},
			Filenames: []string{"*.bug"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if bugsAnalyzerRe.MatchString(text) {
			return 0.7
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (BUGS) Name() string {
	return heartbeat.LanguageBUGS.StringChroma()
}
