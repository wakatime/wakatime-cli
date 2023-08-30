package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Croc lexer.
type Croc struct{}

// Lexer returns the lexer.
func (l Croc) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"croc"},
			Filenames: []string{"*.croc"},
			MimeTypes: []string{"text/x-crocsrc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Croc) Name() string {
	return heartbeat.LanguageCroc.StringChroma()
}
