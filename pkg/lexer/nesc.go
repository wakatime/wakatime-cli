package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// NesC lexer.
type NesC struct{}

// Lexer returns the lexer.
func (l NesC) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"nesc"},
			Filenames: []string{"*.nc"},
			MimeTypes: []string{"text/x-nescsrc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		c := lexers.Get(heartbeat.LanguageC.StringChroma())
		if c == nil {
			return 0
		}

		return c.AnalyseText(text)
	})

	return lexer
}

// Name returns the name of the lexer.
func (NesC) Name() string {
	return heartbeat.LanguageNesC.StringChroma()
}
