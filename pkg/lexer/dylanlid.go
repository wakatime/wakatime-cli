package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// DylanLID lexer.
type DylanLID struct{}

// Lexer returns the lexer.
func (l DylanLID) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"dylan-lid", "lid"},
			Filenames: []string{"*.lid", "*.hdp"},
			MimeTypes: []string{"text/x-dylan-lid"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (DylanLID) Name() string {
	return heartbeat.LanguageDylanLID.StringChroma()
}
