package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var logosAnalyserKeywordsRe = regexp.MustCompile(`%(?:hook|ctor|init|c\()`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageLogos.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"logos"},
			Filenames: []string{"*.x", "*.xi", "*.xm", "*.xmi"},
			MimeTypes: []string{"text/x-logos"},
			Priority:  0.25,
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		if logosAnalyserKeywordsRe.MatchString(text) {
			return 1.0
		}

		return 0
	}))
}
