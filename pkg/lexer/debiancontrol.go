package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// DebianControlFile lexer.
type DebianControlFile struct{}

// Lexer returns the lexer.
func (l DebianControlFile) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"control", "debcontrol"},
			Filenames: []string{"control"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (DebianControlFile) Name() string {
	return heartbeat.LanguageDebianControlFile.StringChroma()
}
