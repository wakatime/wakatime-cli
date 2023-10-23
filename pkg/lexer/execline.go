package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/shebang"

	"github.com/alecthomas/chroma/v2"
)

// Execline lexer.
type Execline struct{}

// Lexer returns the lexer.
func (l Execline) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"execline"},
			Filenames: []string{"*.exec"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if matched, _ := shebang.MatchString(text, "execlineb"); matched {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Execline) Name() string {
	return heartbeat.LanguageExecline.StringChroma()
}
