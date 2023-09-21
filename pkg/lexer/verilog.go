package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageVerilog.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		// Verilog code will use one of reg/wire/assign for sure, and that
		// is not common elsewhere.
		var result float32

		if strings.Contains(text, "reg") {
			result += 0.1
		}

		if strings.Contains(text, "wire") {
			result += 0.1
		}

		if strings.Contains(text, "assign") {
			result += 0.1
		}

		return result
	})
}
