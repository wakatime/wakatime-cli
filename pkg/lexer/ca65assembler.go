package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoglobals
var ca65AnalyserCommentRe = regexp.MustCompile(`(?m)^\s*;`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageCa65Assembler.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"ca65"},
			Filenames: []string{"*.s"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// comments in GAS start with "#".
		if ca65AnalyserCommentRe.MatchString(text) {
			return 0.9
		}

		return 0
	}))
}
