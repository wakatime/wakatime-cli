package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var cbmBasicV2AnalyserRe = regexp.MustCompile(`^\d+`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageCBMBasicV2.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"cbmbas"},
			Filenames: []string{"*.bas"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// if it starts with a line number, it shouldn't be a "modern" Basic
		// like VB.net
		if cbmBasicV2AnalyserRe.MatchString(text) {
			return 0.2
		}

		return 0
	}))
}
