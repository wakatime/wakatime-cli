package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var turtleAnalyserRe = regexp.MustCompile(`^\s*(@base|BASE|@prefix|PREFIX)`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageTurtle.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		// Turtle and Tera Term macro files share the same file extension
		// but each has a recognizable and distinct syntax.
		if turtleAnalyserRe.MatchString(text) {
			return 0.8
		}

		return 0
	})
}
