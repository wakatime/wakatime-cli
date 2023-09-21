package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var tasmAnalyzerRe = regexp.MustCompile(`(?i)PROC`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageTASM.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		if tasmAnalyzerRe.MatchString(text) {
			return 1.0
		}

		return 0
	})
}
