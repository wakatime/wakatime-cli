package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// SublimeTextConfig lexer.
type SublimeTextConfig struct{}

// Lexer returns the lexer.
func (l SublimeTextConfig) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"sublime"},
			Filenames: []string{"*.sublime-settings"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (SublimeTextConfig) Name() string {
	return heartbeat.LanguageSublimeTextConfig.StringChroma()
}
