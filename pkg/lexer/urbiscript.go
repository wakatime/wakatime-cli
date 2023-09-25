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
	language := heartbeat.LanguageUrbiScript.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"urbiscript"},
			Filenames: []string{"*.u"},
			MimeTypes: []string{"application/x-urbiscript"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// This is fairly similar to C and others, but freezeif and
		// waituntil are unique keywords.
		var result float32

		if strings.Contains(text, "freezeif") {
			result += 0.05
		}

		if strings.Contains(text, "waituntil") {
			result += 0.05
		}

		return result
	}))
}
