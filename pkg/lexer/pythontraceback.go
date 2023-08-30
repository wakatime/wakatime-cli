package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// PythonTraceback lexer.
type PythonTraceback struct{}

// Lexer returns the lexer.
func (l PythonTraceback) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"pytb", "py3tb"},
			Filenames: []string{"*.pytb", "*.py3tb"},
			MimeTypes: []string{"text/x-python-traceback", "text/x-python3-traceback"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (PythonTraceback) Name() string {
	return heartbeat.LanguagePythonTraceback.StringChroma()
}
