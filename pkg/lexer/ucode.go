package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Ucode lexer.
type Ucode struct{}

// Lexer returns the lexer.
func (l Ucode) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ucode"},
			Filenames: []string{"*.u", "*.u1", "*.u2"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// endsuspend and endrepeat are unique to this language, and
		// \self, /self doesn't seem to get used anywhere else either.
		var result float32

		if strings.Contains(text, "endsuspend") {
			result += 0.1
		}

		if strings.Contains(text, "endrepeat") {
			result += 0.1
		}

		if strings.Contains(text, ":=") {
			result += 0.01
		}

		if strings.Contains(text, "procedure") && strings.Contains(text, "end") {
			result += 0.01
		}

		// This seems quite unique to unicon -- doesn't appear in any other
		// example source we have (A quick search reveals that \SELF appears in
		// Perl/Raku code)
		if strings.Contains(text, `\self`) && strings.Contains(text, "/self") {
			result += 0.5
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (Ucode) Name() string {
	return heartbeat.LanguageUcode.StringChroma()
}
