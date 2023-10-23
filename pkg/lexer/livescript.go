package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LiveScript lexer.
type LiveScript struct{}

// Lexer returns the lexer.
func (l LiveScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"live-script", "livescript"},
			Filenames: []string{"*.ls"},
			MimeTypes: []string{"text/livescript"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LiveScript) Name() string {
	return heartbeat.LanguageLiveScript.StringChroma()
}
