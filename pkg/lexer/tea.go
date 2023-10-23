package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Tea lexer. Lexer for Tea Templates <http://teatrove.org/>.
type Tea struct{}

// Lexer returns the lexer.
func (l Tea) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"tea"},
			Filenames: []string{"*.tea"},
			MimeTypes: []string{"text/x-tea"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Tea) Name() string {
	return heartbeat.LanguageTea.StringChroma()
}
