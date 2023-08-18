package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// UrbiScript lexer.
type UrbiScript struct{}

// Lexer returns the lexer.
func (l UrbiScript) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"urbiscript"},
			Filenames: []string{"*.u"},
			MimeTypes: []string{"application/x-urbiscript"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// This is fairly similar to C and others, but freezeif and
		// waituntil are unique keywords.
		var result float32

		if strings.Contains(text, "freezeif") {
			result += 0.05
		}

		if strings.Contains(text, "waituntil") {
			result += 0.05
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (UrbiScript) Name() string {
	return heartbeat.LanguageUrbiScript.StringChroma()
}
