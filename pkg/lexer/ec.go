package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// EC lexer.
type EC struct{}

// Lexer returns the lexer.
func (l EC) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ec"},
			Filenames: []string{"*.ec", "*.eh"},
			MimeTypes: []string{"text/x-echdr", "text/x-ecsrc"},
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
func (EC) Name() string {
	return heartbeat.LanguageEC.StringChroma()
}
