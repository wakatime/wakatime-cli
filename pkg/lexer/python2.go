package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/shebang"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Python2 lexer.
type Python2 struct{}

// Lexer returns the lexer.
func (l Python2) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if matched, _ := shebang.MatchString(text, `pythonw?2(\.\d)?`); matched {
				return 1.0
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Python2) Name() string {
	return heartbeat.LanguagePython2.StringChroma()
}
