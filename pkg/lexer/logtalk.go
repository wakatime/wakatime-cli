package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var logtalkAnalyserSyntaxRe = regexp.MustCompile(`(?m)^:-\s[a-z]`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageLogtalk.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"logtalk"},
			Filenames: []string{"*.lgt", "*.logtalk"},
			MimeTypes: []string{"text/x-logtalk"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		if strings.Contains(text, ":- object(") ||
			strings.Contains(text, ":- protocol(") ||
			strings.Contains(text, ":- category(") {
			return 1.0
		}

		if logtalkAnalyserSyntaxRe.MatchString(text) {
			return 0.9
		}

		return 0
	}))
}
