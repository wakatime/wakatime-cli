package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// SARL lexer. For SARL <http://www.sarl.io> source code.
type SARL struct{}

// Lexer returns the lexer.
func (l SARL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"sarl"},
			Filenames: []string{"*.sarl"},
			MimeTypes: []string{"text/x-sarl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (SARL) Name() string {
	return heartbeat.LanguageSARL.StringChroma()
}
