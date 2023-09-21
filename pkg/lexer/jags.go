package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	jagsAnalyserModelRe = regexp.MustCompile(`(?m)^\s*model\s*\{`)
	jagsAnalyserDataRe  = regexp.MustCompile(`(?m)^\s*data\s*\{`)
	jagsAnalyserVarRe   = regexp.MustCompile(`(?m)^\s*var`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageJAGS.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"jags"},
			Filenames: []string{"*.jag", "*.bug"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		if jagsAnalyserModelRe.MatchString(text) {
			if jagsAnalyserDataRe.MatchString(text) {
				return 0.9
			}

			if jagsAnalyserVarRe.MatchString(text) {
				return 0.9
			}

			return 0.3
		}

		return 0
	}))
}
