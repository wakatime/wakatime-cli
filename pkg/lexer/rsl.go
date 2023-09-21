package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var rslAnalyserRe = regexp.MustCompile(`(?i)scheme\s*.*?=\s*class\s*type`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageRSL.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"rsl"},
			Filenames: []string{"*.rsl"},
			MimeTypes: []string{"text/rsl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// Check for the most common text in the beginning of a RSL file.
		if rslAnalyserRe.MatchString(text) {
			return 1.0
		}

		return 0
	}))
}
