package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LiterateHaskell lexer.
type LiterateHaskell struct{}

// Lexer returns the lexer.
func (l LiterateHaskell) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"lhs", "literate-haskell", "lhaskell"},
			Filenames: []string{"*.lhs"},
			MimeTypes: []string{"text/x-literate-haskell"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LiterateHaskell) Name() string {
	return heartbeat.LanguageLiterateHaskell.StringChroma()
}
