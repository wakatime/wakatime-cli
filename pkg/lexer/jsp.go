package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/xml"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageJSP.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"jsp"},
			Filenames: []string{"*.jsp"},
			MimeTypes: []string{"application/x-jsp"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		var result float32

		java := lexers.Get(heartbeat.LanguageJava.StringChroma())
		if java != nil {
			result = java.AnalyseText(text) - 0.01
		}

		if xml.MatchString(text) {
			result += 0.4
		}

		if strings.Contains(text, "<%") && strings.Contains(text, "%>") {
			result += 0.1
		}

		return result
	}))
}
