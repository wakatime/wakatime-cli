package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CsoundDocument lexer.
type CsoundDocument struct{}

// Lexer returns the lexer.
func (l CsoundDocument) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"csound-document", "csound-csd"},
			Filenames: []string{"*.csd"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CsoundDocument) Name() string {
	return heartbeat.LanguageCsoundDocument.StringChroma()
}
