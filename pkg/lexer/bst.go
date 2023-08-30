package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// BST lexer.
type BST struct{}

// Lexer returns the lexer.
func (l BST) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"bst", "bst-pybtex"},
			Filenames: []string{"*.bst"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (BST) Name() string {
	return heartbeat.LanguageBST.StringChroma()
}
