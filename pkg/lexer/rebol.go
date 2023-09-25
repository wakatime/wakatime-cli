package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	rebolAnalyserHeaderRe              = regexp.MustCompile(`^\s*REBOL\s*\[`)
	rebolAnalyserHeaderPrecedingTextRe = regexp.MustCompile(`\s*REBOL\s*\[`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageREBOL.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"rebol"},
			Filenames: []string{"*.r", "*.r3", "*.reb"},
			MimeTypes: []string{"text/x-rebol"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// Check if code contains REBOL header, then it's probably not R code
		if rebolAnalyserHeaderRe.MatchString(text) {
			return 1.0
		}

		if rebolAnalyserHeaderPrecedingTextRe.MatchString(text) {
			return 0.5
		}

		return 0
	}))
}
