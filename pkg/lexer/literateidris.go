package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LiterateIdris lexer.
type LiterateIdris struct{}

// Lexer returns the lexer.
func (l LiterateIdris) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"lidr", "literate-idris", "lidris"},
			Filenames: []string{"*.lidr"},
			MimeTypes: []string{"text/x-literate-idris"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LiterateIdris) Name() string {
	return heartbeat.LanguageLiterateIdris.StringChroma()
}
