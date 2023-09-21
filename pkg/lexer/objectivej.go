package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var objectiveJAnalyserImportRe = regexp.MustCompile(`(?m)^\s*@import\s+[<"]`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageObjectiveJ.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"objective-j", "objectivej", "obj-j", "objj"},
			Filenames: []string{"*.j"},
			MimeTypes: []string{"text/x-objective-j"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// special directive found in most Objective-J files.
		if objectiveJAnalyserImportRe.MatchString(text) {
			return 1.0
		}

		return 0
	}))
}
