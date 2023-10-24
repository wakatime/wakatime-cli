package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Pug lexer.
type Pug struct{}

// Lexer returns the lexer.
func (l Pug) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"pug", "jade"},
			Filenames: []string{"*.pug", "*.jade"},
			MimeTypes: []string{"text/x-pug", "text/x-jade"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Pug) Name() string {
	return heartbeat.LanguagePug.StringChroma()
}
