package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var bugsAnalyzerRe = regexp.MustCompile(`(?m)^\s*model\s*{`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageBUGS.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"bugs", "winbugs", "openbugs"},
			Filenames: []string{"*.bug"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		if bugsAnalyzerRe.MatchString(text) {
			return 0.7
		}

		return 0
	}))
}
