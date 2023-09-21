package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/doctype"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageRHTML.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:           language,
			Aliases:        []string{"rhtml", "html+erb", "html+ruby"},
			Filenames:      []string{"*.rhtml"},
			AliasFilenames: []string{"*.html", "*.htm", "*.xhtml"},
			MimeTypes:      []string{"text/html+ruby"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		erb := lexers.Get(heartbeat.LanguageERB.StringChroma())
		if erb == nil {
			return 0
		}

		result := erb.AnalyseText(text) - 0.01

		if matched, _ := doctype.MatchString(text, "html"); matched {
			// one more than the XmlErbLexer returns
			result += 0.5
		}

		return result
	}))
}
