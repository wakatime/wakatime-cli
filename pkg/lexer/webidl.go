package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// WebIDL lexer.
type WebIDL struct{}

// Lexer returns the lexer.
func (l WebIDL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"webidl"},
			Filenames: []string{"*.webidl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (WebIDL) Name() string {
	return heartbeat.LanguageWebIDL.StringChroma()
}
