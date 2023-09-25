package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var stanAnalyserRe = regexp.MustCompile(`(?m)^\s*parameters\s*\{`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageStan.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"stan"},
			Filenames: []string{"*.stan"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		if stanAnalyserRe.MatchString(text) {
			return 1.0
		}

		return 0
	}))
}
