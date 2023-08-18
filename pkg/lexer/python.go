package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/shebang"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Python lexer.
type Python struct{}

// Lexer returns the lexer.
func (l Python) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
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

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Python) Name() string {
	return heartbeat.LanguagePython.StringChroma()
}
