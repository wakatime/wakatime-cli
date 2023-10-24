package lexer

import (
	"regexp"

	"github.com/alecthomas/chroma/v2"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
)

var nasmAnalyzerRe = regexp.MustCompile(`(?i)PROC`)

// NASM lexer.
type NASM struct{}

// Lexer returns the lexer.
func (l NASM) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			// Probably TASM
			if nasmAnalyzerRe.MatchString(text) {
				return 0
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (NASM) Name() string {
	return heartbeat.LanguageNASM.StringChroma()
}
