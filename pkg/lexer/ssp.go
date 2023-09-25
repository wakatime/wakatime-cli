package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/xml"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var sspAnalyserRe = regexp.MustCompile(`val \w+\s*:`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageSSP.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"ssp"},
			Filenames: []string{"*.ssp"},
			MimeTypes: []string{"application/x-ssp"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		var result float64

		if sspAnalyserRe.MatchString(text) {
			result += 0.6
		}

		if xml.MatchString(text) {
			result += 0.2
		}

		if strings.Contains(text, "<%") && strings.Contains(text, "%>") {
			result += 0.1
		}

		return float32(result)
	}))
}
