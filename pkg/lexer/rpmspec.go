package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RPMSpec lexer.
type RPMSpec struct{}

// Lexer returns the lexer.
func (l RPMSpec) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"spec"},
			Filenames: []string{"*.spec"},
			MimeTypes: []string{"text/x-rpm-spec"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RPMSpec) Name() string {
	return heartbeat.LanguageRPMSpec.StringChroma()
}
