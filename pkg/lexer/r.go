package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/dlclark/regexp2"
)

// nolint:gochecknoglobals
var rAnalyzerRe = regexp2.MustCompile(`[a-z0-9_\])\s]<-(?!-)`, regexp2.None)

// R and also S lexer.
type R struct{}

// Lexer returns the lexer.
func (l R) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			matched, _ := rAnalyzerRe.MatchString(text)
			if matched {
				return 0.11
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (R) Name() string {
	return heartbeat.LanguageR.StringChroma()
}
