package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Cryptol lexer.
type Cryptol struct{}

// Lexer returns the lexer.
func (l Cryptol) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cryptol", "cry"},
			Filenames: []string{"*.cry"},
			MimeTypes: []string{"text/x-cryptol"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Cryptol) Name() string {
	return heartbeat.LanguageCryptol.StringChroma()
}
