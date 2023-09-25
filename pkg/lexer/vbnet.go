package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var vbnetAnalyserRe = regexp.MustCompile(`(?m)^\s*(#If|Module|Namespace)`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageVBNet.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		if vbnetAnalyserRe.MatchString(text) {
			return 0.5
		}

		return 0
	})
}
