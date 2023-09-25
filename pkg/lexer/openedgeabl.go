package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageOpenEdgeABL.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		// try to identify OpenEdge ABL based on a few common constructs.
		var result float32

		if strings.Contains(text, "END.") {
			result += 0.05
		}

		if strings.Contains(text, "END PROCEDURE.") {
			result += 0.05
		}

		if strings.Contains(text, "ELSE DO:") {
			result += 0.05
		}

		return result
	})
}
