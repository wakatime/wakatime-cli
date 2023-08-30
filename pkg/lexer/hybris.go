package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var hybrisAnalyserRe = regexp.MustCompile(`\b(?:public|private)\s+method\b`)

// Hybris lexer.
type Hybris struct{}

// Lexer returns the lexer.
func (l Hybris) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"hybris", "hy"},
			Filenames: []string{"*.hy", "*.hyb"},
			MimeTypes: []string{"text/x-hybris", "application/x-hybris"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// public method and private method don't seem to be quite common
		// elsewhere.
		if hybrisAnalyserRe.MatchString(text) {
			return 0.01
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Hybris) Name() string {
	return heartbeat.LanguageHybris.StringChroma()
}
