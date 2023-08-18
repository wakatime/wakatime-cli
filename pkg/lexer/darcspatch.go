package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// DarcsPatch lexer.
type DarcsPatch struct{}

// Lexer returns the lexer.
func (l DarcsPatch) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"dpatch"},
			Filenames: []string{"*.dpatch", "*.darcspatch"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (DarcsPatch) Name() string {
	return heartbeat.LanguageDarcsPatch.StringChroma()
}
