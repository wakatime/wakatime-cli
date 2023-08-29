package lexer

import (
	"math"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	gapAnalyserDeclarationRe = regexp.MustCompile(
		`(InstallTrueMethod|Declare(Attribute|Category|Filter|Operation|GlobalFunction|Synonym|SynonymAttr|Property))`)
	gapAnalyserImplementationRe = regexp.MustCompile(
		`(DeclareRepresentation|Install(GlobalFunction|Method|ImmediateMethod|OtherMethod)|New(Family|Type)|Objectify)`)
)

// Gap lexer.
type Gap struct{}

// Lexer returns the lexer.
func (l Gap) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"gap"},
			Filenames: []string{"*.g", "*.gd", "*.gi", "*.gap"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		var result float64

		if gapAnalyserDeclarationRe.MatchString(text) {
			result += 0.7
		}

		if gapAnalyserImplementationRe.MatchString(text) {
			result += 0.7
		}

		return float32(math.Min(result, float64(1.0)))
	})

	return lexer
}

// Name returns the name of the lexer.
func (Gap) Name() string {
	return heartbeat.LanguageGap.StringChroma()
}
