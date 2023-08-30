package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var ezhilAnalyserRe = regexp.MustCompile(`[u0b80-u0bff]`)

// Ezhil lexer.
type Ezhil struct{}

// Lexer returns the lexer.
func (l Ezhil) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ezhil"},
			Filenames: []string{"*.n"},
			MimeTypes: []string{"text/x-ezhil"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// this language uses Tamil-script. We'll assume that if there's a
		// decent amount of Tamil-characters, it's this language. This assumption
		// is obviously horribly off if someone uses string literals in tamil
		// in another language.
		if len(ezhilAnalyserRe.FindAllString(text, -1)) > 10 {
			return 0.25
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Ezhil) Name() string {
	return heartbeat.LanguageEzhil.StringChroma()
}
