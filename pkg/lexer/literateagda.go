package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LiterateAgda lexer.
type LiterateAgda struct{}

// Lexer returns the lexer.
func (l LiterateAgda) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"lagda", "literate-agda"},
			Filenames: []string{"*.lagda"},
			MimeTypes: []string{"text/x-literate-agda"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LiterateAgda) Name() string {
	return heartbeat.LanguageLiterateAgda.StringChroma()
}
