package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// NewLisp lexer.
type NewLisp struct{}

// Lexer returns the lexer.
func (l NewLisp) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"newlisp"},
			Filenames: []string{"*.lsp", "*.nl", "*.kif"},
			MimeTypes: []string{"text/x-newlisp", "application/x-newlisp"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (NewLisp) Name() string {
	return heartbeat.LanguageNewLisp.StringChroma()
}
