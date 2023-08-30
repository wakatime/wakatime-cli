package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// ComponentPascal lexer.
type ComponentPascal struct{}

// Lexer returns the lexer.
func (l ComponentPascal) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"componentpascal", "cp"},
			Filenames: []string{"*.cp", "*.cps"},
			MimeTypes: []string{"text/x-component-pascal"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (ComponentPascal) Name() string {
	return heartbeat.LanguageComponentPascal.StringChroma()
}
