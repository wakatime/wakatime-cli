package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	modula2AnalyserProcedureRe = regexp.MustCompile(`\bPROCEDURE\b`)
	modula2AnalyserFunctionRe  = regexp.MustCompile(`\bFUNCTION\b`)
)

// Modula2 lexer.
type Modula2 struct{}

// Lexer returns the lexer.
func (l Modula2) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
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

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Modula2) Name() string {
	return heartbeat.LanguageModula2.StringChroma()
}
