package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// GettextCatalog lexer.
type GettextCatalog struct{}

// Lexer returns the lexer.
func (l GettextCatalog) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"pot", "po"},
			Filenames: []string{"*.pot", "*.po"},
			MimeTypes: []string{"application/x-gettext", "text/x-gettext", "text/gettext"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (GettextCatalog) Name() string {
	return heartbeat.LanguageGettextCatalog.StringChroma()
}
