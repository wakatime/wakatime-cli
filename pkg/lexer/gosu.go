package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Gosu lexer.
type Gosu struct{}

// Lexer returns the lexer.
func (l Gosu) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"gosu"},
			Filenames: []string{"*.gs", "*.gsx", "*.gsp", "*.vark"},
			MimeTypes: []string{"text/x-gosu"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Gosu) Name() string {
	return heartbeat.LanguageGosu.StringChroma()
}
