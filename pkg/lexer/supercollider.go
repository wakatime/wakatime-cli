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
	language := heartbeat.LanguageSuperCollider.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"sc", "supercollider"},
			Filenames: []string{"*.sc", "*.scd"},
			MimeTypes: []string{"application/supercollider", "text/supercollider"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// We're searching for a common function and a unique keyword here.
		if strings.Contains(text, "SinOsc") || strings.Contains(text, "thisFunctionDef") {
			return 0.1
		}

		return 0
	}))
}
