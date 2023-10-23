package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// MozPreprocPercent lexer.
type MozPreprocPercent struct{}

// Lexer returns the lexer.
func (l MozPreprocPercent) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"mozpercentpreproc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (MozPreprocPercent) Name() string {
	return heartbeat.LanguageMozPreprocPercent.StringChroma()
}
