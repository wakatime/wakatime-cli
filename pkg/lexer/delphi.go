package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Delphi lexer.
type Delphi struct{}

// Lexer returns the lexer.
func (l Delphi) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"delphi", "pas", "pascal", "objectpascal"},
			Filenames: []string{"*.pas", "*.dpr"},
			MimeTypes: []string{"text/x-pascal"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Delphi) Name() string {
	return heartbeat.LanguageDelphi.StringChroma()
}
