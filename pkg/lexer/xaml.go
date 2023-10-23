package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// XAML lexer.
type XAML struct{}

// Lexer returns the lexer.
func (l XAML) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"xaml"},
			Filenames: []string{"*.xaml"},
			MimeTypes: []string{"application/xaml+xml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (XAML) Name() string {
	return heartbeat.LanguageXAML.StringChroma()
}
