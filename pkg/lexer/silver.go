package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Silver lexer. For Silver <https://bitbucket.org/viperproject/silver> source code.
type Silver struct{}

// Lexer returns the lexer.
func (l Silver) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"silver"},
			Filenames: []string{"*.sil", "*.vpr"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Silver) Name() string {
	return heartbeat.LanguageSilver.StringChroma()
}
