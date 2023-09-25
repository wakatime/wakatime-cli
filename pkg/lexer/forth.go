package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoglobals
var forthAnalyzerRe = regexp.MustCompile(`\n:[^\n]+;\n`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageForth.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		text = strings.ReplaceAll(text, "\r\n", "\n")

		// Forth uses : COMMAND ; quite a lot in a single line, so we're trying
		// to find that.
		if forthAnalyzerRe.MatchString(text) {
			return 0.3
		}

		return 0
	})
}
