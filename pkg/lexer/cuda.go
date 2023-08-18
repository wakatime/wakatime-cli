package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// CUDA lexer.
type CUDA struct{}

// Lexer returns the lexer.
func (l CUDA) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cuda", "cu"},
			Filenames: []string{"*.cu", "*.cuh"},
			MimeTypes: []string{"text/x-cuda"},
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
func (CUDA) Name() string {
	return heartbeat.LanguageCUDA.StringChroma()
}
