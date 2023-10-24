package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// PyPyLog lexer.
type PyPyLog struct{}

// Lexer returns the lexer.
func (l PyPyLog) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"pypylog", "pypy"},
			Filenames: []string{"*.pypylog"},
			MimeTypes: []string{"application/x-pypylog"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (PyPyLog) Name() string {
	return heartbeat.LanguagePyPyLog.StringChroma()
}
