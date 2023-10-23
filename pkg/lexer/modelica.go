package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Modelica lexer.
type Modelica struct{}

// Lexer returns the lexer.
func (l Modelica) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"modelica"},
			Filenames: []string{"*.mo"},
			MimeTypes: []string{"text/x-modelica"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Modelica) Name() string {
	return heartbeat.LanguageModelica.StringChroma()
}
