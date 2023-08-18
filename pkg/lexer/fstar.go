package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// FStar lexer.
type FStar struct{}

// Lexer returns the lexer.
func (l FStar) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"fstar"},
			Filenames: []string{"*.fst", "*.fsti"},
			MimeTypes: []string{"text/x-fstar"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (FStar) Name() string {
	return heartbeat.LanguageFStar.StringChroma()
}
