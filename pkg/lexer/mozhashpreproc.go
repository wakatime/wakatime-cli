package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MozPreprocHash lexer.
type MozPreprocHash struct{}

// Lexer returns the lexer.
func (l MozPreprocHash) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"mozhashpreproc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MozPreprocHash) Name() string {
	return heartbeat.LanguageMozPreprocHash.StringChroma()
}
