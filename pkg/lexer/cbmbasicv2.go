package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var cbmBasicV2AnalyserRe = regexp.MustCompile(`^\d+`)

// CBMBasicV2 CBM BASIC V2 lexer.
type CBMBasicV2 struct{}

// Lexer returns the lexer.
func (l CBMBasicV2) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cbmbas"},
			Filenames: []string{"*.bas"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// if it starts with a line number, it shouldn't be a "modern" Basic
		// like VB.net
		if cbmBasicV2AnalyserRe.MatchString(text) {
			return 0.2
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (CBMBasicV2) Name() string {
	return heartbeat.LanguageCBMBasicV2.StringChroma()
}
