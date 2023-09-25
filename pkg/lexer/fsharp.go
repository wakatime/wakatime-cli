package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageFSharp.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		// F# doesn't have that many unique features -- |> and <| are weak
		// indicators.
		var result float32

		if strings.Contains(text, "|>") {
			result += 0.05
		}

		if strings.Contains(text, "<|") {
			result += 0.05
		}

		return result
	})
}
