package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/dlclark/regexp2"
)

// nolint:gochecknoglobals
var rAnalyzerRe = regexp2.MustCompile(`[a-z0-9_\])\s]<-(?!-)`, regexp2.None)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageR.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		matched, _ := rAnalyzerRe.MatchString(text)
		if matched {
			return 0.11
		}

		return 0
	})
}
