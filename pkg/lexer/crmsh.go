package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Crmsh lexer.
type Crmsh struct{}

// Lexer returns the lexer.
func (l Crmsh) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"crmsh", "pcmk"},
			Filenames: []string{"*.crmsh", "*.pcmk"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Crmsh) Name() string {
	return heartbeat.LanguageCrmsh.StringChroma()
}
