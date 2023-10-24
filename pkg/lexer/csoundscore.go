package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CsoundScore lexer.
type CsoundScore struct{}

// Lexer returns the lexer.
func (l CsoundScore) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"csound-score", "csound-sco"},
			Filenames: []string{"*.sco"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CsoundScore) Name() string {
	return heartbeat.LanguageCsoundScore.StringChroma()
}
