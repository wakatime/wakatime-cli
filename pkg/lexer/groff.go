package lexer

import (
	"regexp"
	"unicode"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoglobals
var groffAlphanumericRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageGroff.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		if len(text) <= 1 {
			return 0
		}

		if text[:1] != "." {
			return 0
		}

		if len(text) <= 3 {
			return 0
		}

		if text[:3] == `.\"` {
			return 1.0
		}

		if len(text) <= 4 {
			return 0
		}

		if text[:4] == ".TH " {
			return 1.0
		}

		if groffAlphanumericRe.MatchString(text[1:3]) && unicode.IsSpace(rune(text[3])) {
			return 0.9
		}

		return 0
	})
}
