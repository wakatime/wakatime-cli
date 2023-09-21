package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	singularityAnalyserHeaderRe  = regexp.MustCompile(`(?i)\b(?:osversion|includecmd|mirrorurl)\b`)
	singularityAnalyserSectionRe = regexp.MustCompile(
		`%(?:pre|post|setup|environment|help|labels|test|runscript|files|startscript)\b`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageSingularity.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"singularity"},
			Filenames: []string{"*.def", "Singularity"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// This is a quite simple script file, but there are a few keywords
		// which seem unique to this language.
		var result float32

		if singularityAnalyserHeaderRe.MatchString(text) {
			result += 0.5
		}

		if singularityAnalyserSectionRe.MatchString(text) {
			result += 0.49
		}

		return result
	}))
}
