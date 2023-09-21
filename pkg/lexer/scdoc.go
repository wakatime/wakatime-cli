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
	language := heartbeat.LanguageScdoc.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"scdoc", "scd"},
			Filenames: []string{"*.scd", "*.scdoc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// This is very similar to markdown, save for the escape characters
		// needed for * and _.
		var result float32

		if strings.Contains(text, `\*`) {
			result += 0.01
		}

		if strings.Contains(text, `\_`) {
			result += 0.01
		}

		return result
	}))
}
