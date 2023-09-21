package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	modula2AnalyserProcedureRe = regexp.MustCompile(`\bPROCEDURE\b`)
	modula2AnalyserFunctionRe  = regexp.MustCompile(`\bFUNCTION\b`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageModula2.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		// It's Pascal-like, but does not use FUNCTION -- uses PROCEDURE
		// instead.

		// Check if this looks like Pascal, if not, bail out early
		if !strings.Contains(text, "(*") && !strings.Contains(text, "*)") && !strings.Contains(text, ":=") {
			return 0
		}

		var result float32

		// Procedure is in Modula2
		if modula2AnalyserProcedureRe.MatchString(text) {
			result += 0.6
		}

		// FUNCTION is only valid in Pascal, but not in Modula2
		if modula2AnalyserFunctionRe.MatchString(text) {
			result = 0
		}

		return result
	})
}
