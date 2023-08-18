package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// QVTO lexer. For the QVT Operational Mapping language <http://www.omg.org/spec/QVT/1.1/>.
type QVTO struct{}

// Lexer returns the lexer.
func (l QVTO) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"qvto", "qvt"},
			Filenames: []string{"*.qvto"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (QVTO) Name() string {
	return heartbeat.LanguageQVTO.StringChroma()
}
