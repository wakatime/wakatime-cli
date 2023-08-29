package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// VCTreeStatus lexer.
type VCTreeStatus struct{}

// Lexer returns the lexer.
func (l VCTreeStatus) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"vctreestatus"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (VCTreeStatus) Name() string {
	return heartbeat.LanguageVCTreeStatus.StringChroma()
}
