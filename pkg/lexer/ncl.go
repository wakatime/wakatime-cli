package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// NCL lexer.
type NCL struct{}

// Lexer returns the lexer.
func (l NCL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ncl"},
			Filenames: []string{"*.ncl"},
			MimeTypes: []string{"text/ncl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (NCL) Name() string {
	return heartbeat.LanguageNCL.StringChroma()
}
