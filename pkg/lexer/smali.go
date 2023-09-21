package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	smaliAnalyserClassRe         = regexp.MustCompile(`(?m)^\s*\.class\s`)
	smaliAnalyserClassKeywordsRe = regexp.MustCompile(
		`(?m)\b((check-cast|instance-of|throw-verification-error` +
			`)\b|(-to|add|[ais]get|[ais]put|and|cmpl|const|div|` +
			`if|invoke|move|mul|neg|not|or|rem|return|rsub|shl` +
			`|shr|sub|ushr)[-/])|{|}`)
	smaliAnalyserKeywordsRe = regexp.MustCompile(
		`(?m)(\.(catchall|epilogue|restart local|prologue)|` +
			`\b(array-data|class-change-error|declared-synchronized|` +
			`(field|inline|vtable)@0x[0-9a-fA-F]|generic-error|` +
			`illegal-class-access|illegal-field-access|` +
			`illegal-method-access|instantiation-error|no-error|` +
			`no-such-class|no-such-field|no-such-method|` +
			`packed-switch|sparse-switch))\b`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageSmali.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		var result float32

		if smaliAnalyserClassRe.MatchString(text) {
			result += 0.5

			if smaliAnalyserClassKeywordsRe.MatchString(text) {
				result += 0.3
			}
		}

		if smaliAnalyserKeywordsRe.MatchString(text) {
			result += 0.6
		}

		return result
	})
}
