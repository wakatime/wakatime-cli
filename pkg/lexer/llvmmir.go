package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LLVMMIR lexer.
type LLVMMIR struct{}

// Lexer returns the lexer.
func (l LLVMMIR) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"llvm-mir"},
			Filenames: []string{"*.mir"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LLVMMIR) Name() string {
	return heartbeat.LanguageLLVMMIR.StringChroma()
}
