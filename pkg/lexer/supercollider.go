package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// SuperCollider lexer.
type SuperCollider struct{}

// Lexer returns the lexer.
func (l SuperCollider) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"sc", "supercollider"},
			Filenames: []string{"*.sc", "*.scd"},
			MimeTypes: []string{"application/supercollider", "text/supercollider"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// We're searching for a common function and a unique keyword here.
		if strings.Contains(text, "SinOsc") || strings.Contains(text, "thisFunctionDef") {
			return 0.1
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (SuperCollider) Name() string {
	return heartbeat.LanguageSuperCollider.StringChroma()
}
