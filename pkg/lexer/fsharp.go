package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// FSharp lexer.
type FSharp struct{}

// Lexer returns the lexer.
func (l FSharp) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			// F# doesn't have that many unique features -- |> and <| are weak
			// indicators.
			var result float32

			if strings.Contains(text, "|>") {
				result += 0.05
			}

			if strings.Contains(text, "<|") {
				result += 0.05
			}

			return result
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (FSharp) Name() string {
	return heartbeat.LanguageFSharp.StringChroma()
}
