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
	language := heartbeat.LanguageTADS3.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"tads3"},
			Filenames: []string{"*.t"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// This is a rather generic descriptive language without strong
		// identifiers. It looks like a 'GameMainDef' has to be present,
		// and/or a 'versionInfo' with an 'IFID' field.
		var result float32

		if strings.Contains(text, "__TADS") || strings.Contains(text, "GameMainDef") {
			result += 0.2
		}

		// This is a fairly unique keyword which is likely used in source as well.
		if strings.Contains(text, "versionInfo") && strings.Contains(text, "IFID") {
			result += 0.1
		}

		return result
	}))
}
