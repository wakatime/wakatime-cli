package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	gasAnalyzerDirectiveRe      = regexp.MustCompile(`(?m)^\.(text|data|section)`)
	gasAnalyzerOtherDirectiveRe = regexp.MustCompile(`(?m)^\.\w+`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageGas.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		if gasAnalyzerDirectiveRe.MatchString(text) {
			return 1.0
		}

		if gasAnalyzerOtherDirectiveRe.MatchString(text) {
			return 0.1
		}

		return 0
	})
}
