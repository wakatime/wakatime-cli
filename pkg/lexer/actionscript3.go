package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoglobals
var actionscript3AnalyserRe = regexp.MustCompile(`\w+\s*:\s*\w`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageActionScript3.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		if actionscript3AnalyserRe.MatchString(text) {
			return 0.3
		}

		return 0
	})
}
