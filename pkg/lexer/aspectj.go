package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// AspectJ lexer.
type AspectJ struct{}

// Lexer returns the lexer.
func (l AspectJ) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"aspectj"},
			Filenames: []string{"*.aj"},
			MimeTypes: []string{"text/x-aspectj"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (AspectJ) Name() string {
	return heartbeat.LanguageAspectJ.StringChroma()
}
