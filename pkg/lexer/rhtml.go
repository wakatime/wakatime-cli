package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/doctype"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RHTML lexer. Subclass of the ERB lexer that highlights the unlexed data
// with the html lexer.
type RHTML struct{}

// Lexer returns the lexer.
func (l RHTML) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:           l.Name(),
			Aliases:        []string{"rhtml", "html+erb", "html+ruby"},
			Filenames:      []string{"*.rhtml"},
			AliasFilenames: []string{"*.html", "*.htm", "*.xhtml"},
			MimeTypes:      []string{"text/html+ruby"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		result := ERB{}.Lexer().AnalyseText(text) - 0.01

		if matched, _ := doctype.MatchString(text, "html"); matched {
			// one more than the XmlErbLexer returns
			result += 0.5
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (RHTML) Name() string {
	return heartbeat.LanguageRHTML.StringChroma()
}
