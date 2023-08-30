package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// IDL lexer.
type IDL struct{}

// Lexer returns the lexer.
func (l IDL) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"idl"},
			Filenames: []string{"*.pro"},
			MimeTypes: []string{"text/idl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// endelse seems to be unique to IDL, endswitch is rare at least.
		var result float32

		if strings.Contains(text, "endelse") {
			result += 0.2
		}

		if strings.Contains(text, "endswitch") {
			result += 0.01
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (IDL) Name() string {
	return heartbeat.LanguageIDL.StringChroma()
}
