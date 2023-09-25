package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var ezhilAnalyserRe = regexp.MustCompile(`[u0b80-u0bff]`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageEzhil.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"ezhil"},
			Filenames: []string{"*.n"},
			MimeTypes: []string{"text/x-ezhil"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// this language uses Tamil-script. We'll assume that if there's a
		// decent amount of Tamil-characters, it's this language. This assumption
		// is obviously horribly off if someone uses string literals in tamil
		// in another language.
		if len(ezhilAnalyserRe.FindAllString(text, -1)) > 10 {
			return 0.25
		}

		return 0
	}))
}
