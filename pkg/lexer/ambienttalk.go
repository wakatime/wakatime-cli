package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// AmbientTalk lexer.
type AmbientTalk struct{}

// Lexer returns the lexer.
func (l AmbientTalk) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"at", "ambienttalk", "ambienttalk/2"},
			Filenames: []string{"*.at"},
			MimeTypes: []string{"text/x-ambienttalk"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (AmbientTalk) Name() string {
	return heartbeat.LanguageAmbientTalk.StringChroma()
}
