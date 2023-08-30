package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoglobals
var forthAnalyzerRe = regexp.MustCompile(`\n:[^\n]+;\n`)

// Forth lexer.
type Forth struct{}

// Lexer returns the lexer.
func (l Forth) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			text = strings.ReplaceAll(text, "\r\n", "\n")

			// Forth uses : COMMAND ; quite a lot in a single line, so we're trying
			// to find that.
			if forthAnalyzerRe.MatchString(text) {
				return 0.3
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Forth) Name() string {
	return heartbeat.LanguageForth.StringChroma()
}
