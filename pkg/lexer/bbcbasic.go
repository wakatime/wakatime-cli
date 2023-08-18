package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// BBCBasic lexer.
type BBCBasic struct{}

// Lexer returns the lexer.
func (l BBCBasic) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"bbcbasic"},
			Filenames: []string{"*.bbc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if strings.HasPrefix(text, "10REM >") || strings.HasPrefix(text, "REM >") {
			return 0.9
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (BBCBasic) Name() string {
	return heartbeat.LanguageBBCBasic.StringChroma()
}
