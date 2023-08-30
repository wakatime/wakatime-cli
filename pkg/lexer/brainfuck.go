package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Brainfuck lexer.
type Brainfuck struct{}

// Lexer returns the lexer.
func (l Brainfuck) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	var (
		ok       bool
		rgxlexer *chroma.RegexLexer
	)

	if rgxlexer, ok = lexer.(*chroma.RegexLexer); !ok {
		return nil
	}

	rgxlexer.SetAnalyser(func(text string) float32 {
		// it's safe to assume that a program which mostly consists of + -
		// and < > is brainfuck.
		var plusMinusCount float64
		var greaterLessCount float64

		rangeToCheck := len(text)

		if rangeToCheck > 256 {
			rangeToCheck = 256
		}

		for _, c := range text[:rangeToCheck] {
			if c == '+' || c == '-' {
				plusMinusCount++
			}
			if c == '<' || c == '>' {
				greaterLessCount++
			}
		}

		if plusMinusCount > (0.25 * float64(rangeToCheck)) {
			return 1.0
		}

		if greaterLessCount > (0.25 * float64(rangeToCheck)) {
			return 1.0
		}

		if strings.Contains(text, "[-]") {
			return 0.5
		}

		return 0
	})

	return rgxlexer
}

// Name returns the name of the lexer.
func (Brainfuck) Name() string {
	return heartbeat.LanguageBrainfuck.StringChroma()
}
