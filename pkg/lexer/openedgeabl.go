package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// OpenEdgeABL lexer.
type OpenEdgeABL struct{}

// Lexer returns the lexer.
func (l OpenEdgeABL) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
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

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (OpenEdgeABL) Name() string {
	return heartbeat.LanguageOpenEdgeABL.StringChroma()
}
