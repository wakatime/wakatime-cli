package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var jclAnalyserJobHeaderRe = regexp.MustCompile(`(?i)^//[a-z#$@][a-z0-9#$@]{0,7}\s+job(\s+.*)?$`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageJCL.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"jcl"},
			Filenames: []string{"*.jcl"},
			MimeTypes: []string{"text/x-jcl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// Recognize JCL job by header.
		lines := strings.Split(text, "\n")
		if len(lines) == 0 {
			return 0
		}

		if jclAnalyserJobHeaderRe.MatchString(lines[0]) {
			return 1.0
		}

		return 0
	}))
}
