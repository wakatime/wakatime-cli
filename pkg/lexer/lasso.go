package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	lassoAnalyserDelimiterRe = regexp.MustCompile(`(?i)<\?lasso`)
	lassoAnalyserLocalRe     = regexp.MustCompile(`(?i)local\(`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageLasso.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:           language,
			Aliases:        []string{"lasso", "lassoscript"},
			Filenames:      []string{"*.lasso", "*.lasso[89]"},
			AliasFilenames: []string{"*.incl", "*.inc", "*.las"},
			MimeTypes:      []string{"text/x-lasso"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		var result float32

		if strings.Contains(text, "bin/lasso9") {
			result += 0.8
		}

		if lassoAnalyserDelimiterRe.MatchString(text) {
			result += 0.4
		}

		if lassoAnalyserLocalRe.MatchString(text) {
			result += 0.4
		}

		return result
	}))
}
