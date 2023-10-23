package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var logtalkAnalyserSyntaxRe = regexp.MustCompile(`(?m)^:-\s[a-z]`)

// Logtalk lexer.
type Logtalk struct{}

// Lexer returns the lexer.
func (l Logtalk) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"logtalk"},
			Filenames: []string{"*.lgt", "*.logtalk"},
			MimeTypes: []string{"text/x-logtalk"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if strings.Contains(text, ":- object(") ||
			strings.Contains(text, ":- protocol(") ||
			strings.Contains(text, ":- category(") {
			return 1.0
		}

		if logtalkAnalyserSyntaxRe.MatchString(text) {
			return 0.9
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Logtalk) Name() string {
	return heartbeat.LanguageLogtalk.StringChroma()
}
