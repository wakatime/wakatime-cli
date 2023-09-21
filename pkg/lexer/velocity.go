package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	velocityAnalzserMacroRe     = regexp.MustCompile(`(?s)#\{?macro\}?\(.*?\).*?#\{?end\}?`)
	velocityAnalzserIfRe        = regexp.MustCompile(`(?s)#\{?if\}?\(.+?\).*?#\{?end\}?`)
	velocityAnalzserForeachRe   = regexp.MustCompile(`(?s)#\{?foreach\}?\(.+?\).*?#\{?end\}?`)
	velocityAnalzserReferenceRe = regexp.MustCompile(`\$!?\{?[a-zA-Z_]\w*(\([^)]*\))?(\.\w+(\([^)]*\))?)*\}?`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageVelocity.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"velocity"},
			Filenames: []string{"*.vm", "*.fhtml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		var result float64

		if velocityAnalzserMacroRe.MatchString(text) {
			result += 0.25
		}

		if velocityAnalzserIfRe.MatchString(text) {
			result += 0.15
		}

		if velocityAnalzserForeachRe.MatchString(text) {
			result += 0.15
		}

		if velocityAnalzserReferenceRe.MatchString(text) {
			result += 0.01
		}

		return float32(result)
	}))
}
