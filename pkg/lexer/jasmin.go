package lexer

import (
	"math"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	jasminAnalyserClassRe       = regexp.MustCompile(`(?m)^\s*\.class\s`)
	jasminAnalyserInstructionRe = regexp.MustCompile(`(?m)^\s*[a-z]+_[a-z]+\b`)
	jasminAnalyserKeywordsRe    = regexp.MustCompile(
		`(?m)^\s*\.(attribute|bytecode|debug|deprecated|enclosing|inner|interface|limit|set|signature|stack)\b`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageJasmin.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"jasmin", "jasminxt"},
			Filenames: []string{"*.j"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
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
	}))
}
