package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/xml"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// JSP lexer.
type JSP struct{}

// Lexer returns the lexer.
func (l JSP) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"jsp"},
			Filenames: []string{"*.jsp"},
			MimeTypes: []string{"application/x-jsp"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		var result float32

		java := lexers.Get(heartbeat.LanguageJava.StringChroma())
		if java != nil {
			result = java.AnalyseText(text) - 0.01
		}

		if xml.MatchString(text) {
			result += 0.4
		}

		if strings.Contains(text, "<%") && strings.Contains(text, "%>") {
			result += 0.1
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (JSP) Name() string {
	return heartbeat.LanguageJSP.StringChroma()
}
