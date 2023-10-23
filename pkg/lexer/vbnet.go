package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var vbnetAnalyserRe = regexp.MustCompile(`(?m)^\s*(#If|Module|Namespace)`)

// VBNet lexer.
type VBNet struct{}

// Lexer returns the lexer.
func (l VBNet) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if vbnetAnalyserRe.MatchString(text) {
				return 0.5
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (VBNet) Name() string {
	return heartbeat.LanguageVBNet.StringChroma()
}
