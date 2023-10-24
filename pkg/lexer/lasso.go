package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	lassoAnalyserDelimiterRe = regexp.MustCompile(`(?i)<\?lasso`)
	lassoAnalyserLocalRe     = regexp.MustCompile(`(?i)local\(`)
)

// Lasso lexer.
type Lasso struct{}

// Lexer returns the lexer.
func (l Lasso) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name: l.Name(),
			Aliases: []string{
				"lasso",
				"lassoscript",
			},
			Filenames: []string{
				"*.lasso",
				"*.lasso[89]",
			},
			AliasFilenames: []string{
				"*.incl",
				"*.inc",
				"*.las",
			},
			MimeTypes: []string{"text/x-lasso"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		var result float32

		if strings.Contains(text, "bin/lasso9") {
			result += 0.8
		}

		if lassoAnalyserDelimiterRe.MatchString(text) {
			result += 0.4
		}

		if lassoAnalyserLocalRe.MatchString(text) {
			result += 0.4
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (Lasso) Name() string {
	return heartbeat.LanguageLasso.StringChroma()
}
