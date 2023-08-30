package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Treetop lexer. A lexer for Treetop <http://treetop.rubyforge.org/> grammars.
type Treetop struct{}

// Lexer returns the lexer.
func (l Treetop) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"treetop"},
			Filenames: []string{"*.treetop", "*.tt"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Treetop) Name() string {
	return heartbeat.LanguageTreetop.StringChroma()
}
