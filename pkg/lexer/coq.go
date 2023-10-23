package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// Coq lexer.
type Coq struct{}

// Lexer returns the lexer.
func (l Coq) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if strings.Contains(text, "Qed") && strings.Contains(text, "Proof") {
				return 1.0
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Coq) Name() string {
	return heartbeat.LanguageCoq.StringChroma()
}
