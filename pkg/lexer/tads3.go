package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// TADS3 lexer.
type TADS3 struct{}

// Lexer returns the lexer.
func (l TADS3) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"tads3"},
			Filenames: []string{"*.t"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// This is a rather generic descriptive language without strong
		// identifiers. It looks like a 'GameMainDef' has to be present,
		// and/or a 'versionInfo' with an 'IFID' field.
		var result float32

		if strings.Contains(text, "__TADS") || strings.Contains(text, "GameMainDef") {
			result += 0.2
		}

		// This is a fairly unique keyword which is likely used in source as well.
		if strings.Contains(text, "versionInfo") && strings.Contains(text, "IFID") {
			result += 0.1
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (TADS3) Name() string {
	return heartbeat.LanguageTADS3.StringChroma()
}
