package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var inform6AnalyserRe = regexp.MustCompile(`(?i)\borigsource\b`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageInform6.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"inform6", "i6"},
			Filenames: []string{"*.inf"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// We try to find a keyword which seem relatively common, unfortunately
		// there is a decent overlap with Smalltalk keywords otherwise here.
		if inform6AnalyserRe.MatchString(text) {
			return 0.05
		}

		return 0
	}))
}
