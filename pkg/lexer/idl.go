package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageIDL.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"idl"},
			Filenames: []string{"*.pro"},
			MimeTypes: []string{"text/idl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// endelse seems to be unique to IDL, endswitch is rare at least.
		var result float32

		if strings.Contains(text, "endelse") {
			result += 0.2
		}

		if strings.Contains(text, "endswitch") {
			result += 0.01
		}

		return result
	}))
}
