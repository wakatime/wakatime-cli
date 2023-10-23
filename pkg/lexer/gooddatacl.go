package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// GoodDataCL lexer.
type GoodDataCL struct{}

// Lexer returns the lexer.
func (l GoodDataCL) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"gooddata-cl"},
			Filenames: []string{"*.gdc"},
			MimeTypes: []string{"text/x-gooddata-cl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (GoodDataCL) Name() string {
	return heartbeat.LanguageGoodDataCL.StringChroma()
}
