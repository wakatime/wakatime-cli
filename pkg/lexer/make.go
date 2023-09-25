package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var makefileAnalyserVariableRe = regexp.MustCompile(`\$\([A-Z_]+\)`)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageMakefile.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		// Many makefiles have $(BIG_CAPS) style variables.
		if makefileAnalyserVariableRe.MatchString(text) {
			return 0.1
		}

		return 0
	})
}
