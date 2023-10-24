package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Verilog lexer.
type Verilog struct{}

// Lexer returns the lexer.
func (l Verilog) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			// Verilog code will use one of reg/wire/assign for sure, and that
			// is not common elsewhere.
			var result float32

			if strings.Contains(text, "reg") {
				result += 0.1
			}

			if strings.Contains(text, "wire") {
				result += 0.1
			}

			if strings.Contains(text, "assign") {
				result += 0.1
			}

			return result
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Verilog) Name() string {
	return heartbeat.LanguageVerilog.StringChroma()
}
