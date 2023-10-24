package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// Python2Traceback lexer.
type Python2Traceback struct{}

// Lexer returns the lexer.
func (l Python2Traceback) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"py2tb"},
			Filenames: []string{"*.py2tb"},
			MimeTypes: []string{"text/x-python2-traceback"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (Python2Traceback) Name() string {
	return heartbeat.LanguagePython2Traceback.StringChroma()
}
