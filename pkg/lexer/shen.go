package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Shen lexer. Lexer for Shen <http://shenlanguage.org/> source code.
type Shen struct{}

// Lexer returns the lexer.
func (l Shen) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"shen"},
			Filenames: []string{"*.shen"},
			MimeTypes: []string{"text/x-shen", "application/x-shen"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Shen) Name() string {
	return heartbeat.LanguageShen.StringChroma()
}
