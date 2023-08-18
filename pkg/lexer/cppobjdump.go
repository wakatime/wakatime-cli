package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// CppObjdump lexer.
type CppObjdump struct{}

// Lexer returns the lexer.
func (l CppObjdump) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"cpp-objdump", "c++-objdumb", "cxx-objdump"},
			Filenames: []string{"*.cpp-objdump", "*.c++-objdump", "*.cxx-objdump"},
			MimeTypes: []string{"text/x-cpp-objdump"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (CppObjdump) Name() string {
	return heartbeat.LanguageCppObjdump.StringChroma()
}
