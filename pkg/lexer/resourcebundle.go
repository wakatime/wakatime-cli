package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ResourceBundle lexer. Lexer for ICU ResourceBundle bundles
// <http://userguide.icu-project.org/locale/resources>
type ResourceBundle struct{}

// Lexer returns the lexer.
func (l ResourceBundle) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"resource", "resourcebundle"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if strings.HasPrefix(text, "root:table") {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (ResourceBundle) Name() string {
	return heartbeat.LanguageResourceBundle.StringChroma()
}
