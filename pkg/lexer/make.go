package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var makefileAnalyserVariableRe = regexp.MustCompile(`\$\([A-Z_]+\)`)

// Makefile lexer.
type Makefile struct{}

// Lexer returns the lexer.
func (l Makefile) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			// Many makefiles have $(BIG_CAPS) style variables.
			if makefileAnalyserVariableRe.MatchString(text) {
				return 0.1
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Makefile) Name() string {
	return heartbeat.LanguageMakefile.StringChroma()
}
