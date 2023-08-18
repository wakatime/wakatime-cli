package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// LLVMMIRBODY lexer.
type LLVMMIRBODY struct{}

// Lexer returns the lexer.
func (l LLVMMIRBODY) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"llvm-mir-body"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (LLVMMIRBODY) Name() string {
	return heartbeat.LanguageLLVMMIRBody.StringChroma()
}
