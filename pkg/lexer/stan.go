package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var stanAnalyserRe = regexp.MustCompile(`(?m)^\s*parameters\s*\{`)

// Stan lexer. Lexer for Stan models.
//
// The Stan modeling language is specified in the *Stan Modeling Language
// User's Guide and Reference Manual, v2.17.0*,
// pdf <https://github.com/stan-dev/stan/releases/download/v2.17.0/stan-reference-2.17.0.pdf>`.
type Stan struct{}

// Lexer returns the lexer.
func (l Stan) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"stan"},
			Filenames: []string{"*.stan"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if stanAnalyserRe.MatchString(text) {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Stan) Name() string {
	return heartbeat.LanguageStan.StringChroma()
}
