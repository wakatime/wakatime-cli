package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// VCLSnippets lexer.
type VCLSnippets struct{}

// Lexer returns the lexer.
func (l VCLSnippets) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"vclsnippets", "vclsnippet"},
			MimeTypes: []string{"text/x-vclsnippet"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (VCLSnippets) Name() string {
	return heartbeat.LanguageVCLSnippets.StringChroma()
}
