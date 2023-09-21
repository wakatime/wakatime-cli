package lexer

import (
	"math"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	gapAnalyserDeclarationRe = regexp.MustCompile(
		`(InstallTrueMethod|Declare(Attribute|Category|Filter|Operation|GlobalFunction|Synonym|SynonymAttr|Property))`)
	gapAnalyserImplementationRe = regexp.MustCompile(
		`(DeclareRepresentation|Install(GlobalFunction|Method|ImmediateMethod|OtherMethod)|New(Family|Type)|Objectify)`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageGap.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"gap"},
			Filenames: []string{"*.g", "*.gd", "*.gi", "*.gap"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		var result float64

		if gapAnalyserDeclarationRe.MatchString(text) {
			result += 0.7
		}

		if gapAnalyserImplementationRe.MatchString(text) {
			result += 0.7
		}

		return float32(math.Min(result, float64(1.0)))
	}))
}
