package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Eiffel lexer.
type Eiffel struct{}

// Lexer returns the lexer.
func (l Eiffel) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"eiffel"},
			Filenames: []string{"*.e"},
			MimeTypes: []string{"text/x-eiffel"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Eiffel) Name() string {
	return heartbeat.LanguageEiffel.StringChroma()
}
