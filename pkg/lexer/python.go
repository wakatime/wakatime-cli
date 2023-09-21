package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/shebang"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguagePython.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		matched, _ := shebang.MatchString(text, `pythonw?(3(\.\d)?)?`)

		if len(text) > 1000 {
			text = text[:1000]
		}

		if matched || strings.Contains(text, "import ") {
			return 1.0
		}

		return 0
	})
}
