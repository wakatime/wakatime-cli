package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageBrainfuck.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		// it's safe to assume that a program which mostly consists of + -
		// and < > is brainfuck.
		var plusMinusCount float64
		var greaterLessCount float64

		rangeToCheck := len(text)

		if rangeToCheck > 256 {
			rangeToCheck = 256
		}

		for _, c := range text[:rangeToCheck] {
			if c == '+' || c == '-' {
				plusMinusCount++
			}
			if c == '<' || c == '>' {
				greaterLessCount++
			}
		}

		if plusMinusCount > (0.25 * float64(rangeToCheck)) {
			return 1.0
		}

		if greaterLessCount > (0.25 * float64(rangeToCheck)) {
			return 1.0
		}

		if strings.Contains(text, "[-]") {
			return 0.5
		}

		return 0
	})
}
