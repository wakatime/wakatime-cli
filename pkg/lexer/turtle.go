package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var turtleAnalyserRe = regexp.MustCompile(`^\s*(@base|BASE|@prefix|PREFIX)`)

// Turtle lexer.
type Turtle struct{}

// Lexer returns the lexer.
func (l Turtle) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			// Turtle and Tera Term macro files share the same file extension
			// but each has a recognizable and distinct syntax.
			if turtleAnalyserRe.MatchString(text) {
				return 0.8
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Turtle) Name() string {
	return heartbeat.LanguageTurtle.StringChroma()
}
