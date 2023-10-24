package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RQL lexer for Relation Query Language <http://www.logilab.org/project/rql>
type RQL struct{}

// Lexer returns the lexer.
func (l RQL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rql"},
			Filenames: []string{"*.rql"},
			MimeTypes: []string{"text/x-rql"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RQL) Name() string {
	return heartbeat.LanguageRQL.StringChroma()
}
