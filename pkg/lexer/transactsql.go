package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	tSQLAnalyserGoRe                  = regexp.MustCompile(`(?i)\bgo\b`)
	tSQLAnalyserDeclareRe             = regexp.MustCompile(`(?i)\bdeclare\s+@`)
	tSQLAnalyserVariableRe            = regexp.MustCompile(`@[a-zA-Z_]\w*\b`)
	tSQLAnalyserNameBetweenBacktickRe = regexp.MustCompile("`[a-zA-Z_]\\w*`")
	tSQLAnalyserNameBetweenBracketRe  = regexp.MustCompile(`\[[a-zA-Z_]\w*\]`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageTransactSQL.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		if tSQLAnalyserDeclareRe.MatchString(text) {
			// Found T-SQL variable declaration.
			return 1.0
		}

		nameBetweenBacktickCount := len(tSQLAnalyserNameBetweenBacktickRe.FindAllString(text, -1))
		nameBetweenBracketCount := len(tSQLAnalyserNameBetweenBracketRe.FindAllString(text, -1))

		var result float32

		// We need to check if there are any names using
		// backticks or brackets, as otherwise both are 0
		// and 0 >= 2 * 0, so we would always assume it's true
		dialectNameCount := nameBetweenBacktickCount + nameBetweenBracketCount

		// nolint: gocritic
		if dialectNameCount >= 1 && nameBetweenBracketCount >= (2*nameBetweenBacktickCount) {
			// Found at least twice as many [name] as `name`.
			result += 0.5
		} else if nameBetweenBracketCount > nameBetweenBacktickCount {
			result += 0.2
		} else if nameBetweenBracketCount > 0 {
			result += 0.1
		}

		if tSQLAnalyserVariableRe.MatchString(text) {
			result += 0.1
		}

		if tSQLAnalyserGoRe.MatchString(text) {
			result += 0.1
		}

		return result
	})
}
