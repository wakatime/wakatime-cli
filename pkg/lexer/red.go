package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Red lexer. A Red-language <http://www.red-lang.org/> lexer.
type Red struct{}

// Lexer returns the lexer.
func (l Red) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"red", "red/system"},
			Filenames: []string{"*.red", "*.reds"},
			MimeTypes: []string{"text/x-red", "text/x-red-system"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Red) Name() string {
	return heartbeat.LanguageRed.StringChroma()
}
