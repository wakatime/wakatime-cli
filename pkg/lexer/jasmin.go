package lexer

import (
	"math"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	jasminAnalyserClassRe       = regexp.MustCompile(`(?m)^\s*\.class\s`)
	jasminAnalyserInstructionRe = regexp.MustCompile(`(?m)^\s*[a-z]+_[a-z]+\b`)
	jasminAnalyserKeywordsRe    = regexp.MustCompile(
		`(?m)^\s*\.(attribute|bytecode|debug|deprecated|enclosing|inner|interface|limit|set|signature|stack)\b`)
)

// Jasmin lexer.
type Jasmin struct{}

// Lexer returns the lexer.
func (l Jasmin) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"jasmin", "jasminxt"},
			Filenames: []string{"*.j"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		var result float64

		if jasminAnalyserClassRe.MatchString(text) {
			result += 0.5

			if jasminAnalyserInstructionRe.MatchString(text) {
				result += 0.3
			}
		}

		if jasminAnalyserKeywordsRe.MatchString(text) {
			result += 0.6
		}

		return float32(math.Min(result, float64(1.0)))
	})

	return lexer
}

// Name returns the name of the lexer.
func (Jasmin) Name() string {
	return heartbeat.LanguageJasmin.StringChroma()
}
