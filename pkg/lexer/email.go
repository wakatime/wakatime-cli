package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// EMail lexer.
type EMail struct{}

// Lexer returns the lexer.
func (l EMail) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"email", "eml"},
			Filenames: []string{"*.eml"},
			MimeTypes: []string{"message/rfc822"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (EMail) Name() string {
	return heartbeat.LanguageEMail.StringChroma()
}
