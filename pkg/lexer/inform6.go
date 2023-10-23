package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var inform6AnalyserRe = regexp.MustCompile(`(?i)\borigsource\b`)

// Inform6 lexer.
type Inform6 struct{}

// Lexer returns the lexer.
func (l Inform6) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"inform6", "i6"},
			Filenames: []string{"*.inf"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// We try to find a keyword which seem relatively common, unfortunately
		// there is a decent overlap with Smalltalk keywords otherwise here.
		if inform6AnalyserRe.MatchString(text) {
			return 0.05
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Inform6) Name() string {
	return heartbeat.LanguageInform6.StringChroma()
}
