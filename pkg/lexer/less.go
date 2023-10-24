package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Less lexer.
type Less struct{}

// Lexer returns the lexer.
func (l Less) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"less"},
			Filenames: []string{"*.less"},
			MimeTypes: []string{"text/x-less-css"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Less) Name() string {
	return heartbeat.LanguageLess.StringChroma()
}
