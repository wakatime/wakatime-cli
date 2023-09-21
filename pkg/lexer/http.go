package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageHTTP.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexers.HTTP.SetAnalyser(func(text string) float32 {
		if strings.HasPrefix(text, "GET") ||
			strings.HasPrefix(text, "POST") ||
			strings.HasPrefix(text, "PUT") ||
			strings.HasPrefix(text, "DELETE") ||
			strings.HasPrefix(text, "HEAD") ||
			strings.HasPrefix(text, "OPTIONS") ||
			strings.HasPrefix(text, "TRACE") ||
			strings.HasPrefix(text, "PATCH") {
			return 1.0
		}

		return 0
	})
}
