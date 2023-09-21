package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var hybrisAnalyserRe = regexp.MustCompile(`\b(?:public|private)\s+method\b`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageHybris.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"hybris", "hy"},
			Filenames: []string{"*.hy", "*.hyb"},
			MimeTypes: []string{"text/x-hybris", "application/x-hybris"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// public method and private method don't seem to be quite common
		// elsewhere.
		if hybrisAnalyserRe.MatchString(text) {
			return 0.01
		}

		return 0
	}))
}
