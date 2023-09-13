package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// IDA lexer.
type IDA struct{}

// Lexer returns the lexer.
func (l IDA) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"IDA Pro", "IDA Free"},
			Filenames: []string{"*.i64", "*.idb"},
			MimeTypes: []string{"text/x-ida"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (IDA) Name() string {
	return heartbeat.LanguageIDA.StringChroma()
}
