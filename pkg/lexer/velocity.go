package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	velocityAnalzserMacroRe     = regexp.MustCompile(`(?s)#\{?macro\}?\(.*?\).*?#\{?end\}?`)
	velocityAnalzserIfRe        = regexp.MustCompile(`(?s)#\{?if\}?\(.+?\).*?#\{?end\}?`)
	velocityAnalzserForeachRe   = regexp.MustCompile(`(?s)#\{?foreach\}?\(.+?\).*?#\{?end\}?`)
	velocityAnalzserReferenceRe = regexp.MustCompile(`\$!?\{?[a-zA-Z_]\w*(\([^)]*\))?(\.\w+(\([^)]*\))?)*\}?`)
)

// Velocity lexer.
type Velocity struct{}

// Lexer returns the lexer.
func (l Velocity) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"velocity"},
			Filenames: []string{"*.vm", "*.fhtml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		var result float64

		if velocityAnalzserMacroRe.MatchString(text) {
			result += 0.25
		}

		if velocityAnalzserIfRe.MatchString(text) {
			result += 0.15
		}

		if velocityAnalzserForeachRe.MatchString(text) {
			result += 0.15
		}

		if velocityAnalzserReferenceRe.MatchString(text) {
			result += 0.01
		}

		return float32(result)
	})

	return lexer
}

// Name returns the name of the lexer.
func (Velocity) Name() string {
	return heartbeat.LanguageVelocity.StringChroma()
}
