package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// RConsole lexer. For R console transcripts or R CMD BATCH output files.
type RConsole struct{}

// Lexer returns the lexer.
func (l RConsole) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rconsole", "rout"},
			Filenames: []string{"*.Rout"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (RConsole) Name() string {
	return heartbeat.LanguageRConsole.StringChroma()
}
