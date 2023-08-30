package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RagelEmbedded lexer. A lexer for Ragel embedded in a host language file.
type RagelEmbedded struct{}

// Lexer returns the lexer.
func (l RagelEmbedded) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ragel-em"},
			Filenames: []string{"*.rl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if strings.Contains(text, "@LANG: indep") {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (RagelEmbedded) Name() string {
	return heartbeat.LanguageRagelEmbedded.StringChroma()
}
