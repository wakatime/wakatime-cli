package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Pawn lexer.
type Pawn struct{}

// Lexer returns the lexer.
func (l Pawn) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"pawn"},
			Filenames: []string{"*.p", "*.pwn", "*.inc"},
			MimeTypes: []string{"text/x-pawn"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// This is basically C. There is a keyword which doesn't exist in C
		// though and is nearly unique to this language.
		if strings.Contains(text, "tagof") {
			return 0.01
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Pawn) Name() string {
	return heartbeat.LanguagePawn.StringChroma()
}
