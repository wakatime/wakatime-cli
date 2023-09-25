package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var limboAnalyzerRe = regexp.MustCompile(`(?m)^implement \w+;`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageLimbo.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"limbo"},
			Filenames: []string{"*.b"},
			MimeTypes: []string{"text/limbo"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// Any limbo module implements something
		if limboAnalyzerRe.MatchString(text) {
			return 0.7
		}

		return 0
	}))
}
