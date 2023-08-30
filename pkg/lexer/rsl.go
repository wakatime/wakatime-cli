package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var rslAnalyserRe = regexp.MustCompile(`(?i)scheme\s*.*?=\s*class\s*type`)

// RSL lexer. RSL <http://en.wikipedia.org/wiki/RAISE> is the formal
// specification language used in RAISE (Rigorous Approach to Industrial
// Software Engineering) method.
type RSL struct{}

// Lexer returns the lexer.
func (l RSL) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rsl"},
			Filenames: []string{"*.rsl"},
			MimeTypes: []string{"text/rsl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// Check for the most common text in the beginning of a RSL file.
		if rslAnalyserRe.MatchString(text) {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (RSL) Name() string {
	return heartbeat.LanguageRSL.StringChroma()
}
