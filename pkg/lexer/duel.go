package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Duel lexer.
type Duel struct{}

// Lexer returns the lexer.
func (l Duel) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"duel", "jbst", "jsonml+bst"},
			Filenames: []string{"*.duel", "*.jbst"},
			MimeTypes: []string{"text/x-duel", "text/x-jbst"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Duel) Name() string {
	return heartbeat.LanguageDuel.StringChroma()
}
