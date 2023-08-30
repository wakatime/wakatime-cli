package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// PEG lexer.
type PEG struct{}

// Lexer returns the lexer.
func (l PEG) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"peg"},
			Filenames: []string{"*.peg"},
			MimeTypes: []string{"text/x-peg"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (PEG) Name() string {
	return heartbeat.LanguagePEG.StringChroma()
}
