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
	language := heartbeat.LanguageECL.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"ecl"},
			Filenames: []string{"*.ecl"},
			MimeTypes: []string{"application/x-ecl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// This is very difficult to guess relative to other business languages.
		// -> in conjunction with BEGIN/END seems relatively rare though.

		var result float32

		if strings.Contains(text, "->") {
			result += 0.01
		}

		if strings.Contains(text, "BEGIN") {
			result += 0.01
		}

		if strings.Contains(text, "END") {
			result += 0.01
		}

		return result
	}))
}
