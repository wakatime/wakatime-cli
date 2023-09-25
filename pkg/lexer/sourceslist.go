package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var sourcesListAnalyserRe = regexp.MustCompile(`(?m)^\s*(deb|deb-src) `)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageSourcesList.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"sourceslist", "sources.list", "debsources"},
			Filenames: []string{"sources.list"},
			MimeTypes: []string{"application/x-debian-sourceslist"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		if sourcesListAnalyserRe.MatchString(text) {
			return 1.0
		}

		return 0
	}))
}
