package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// VCL lexer.
type VCL struct{}

// Lexer returns the lexer.
func (l VCL) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"vcl"},
			Filenames: []string{"*.vcl"},
			MimeTypes: []string{"text/x-vclsrc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// If the very first line is 'vcl 4.0;' it's pretty much guaranteed
		// that this is VCL
		if strings.HasPrefix(text, "vcl 4.0;") {
			return 1.0
		}

		if len(text) > 1000 {
			text = text[:1000]
		}

		// Skip over comments and blank lines
		// This is accurate enough that returning 0.9 is reasonable.
		// Almost no VCL files start without some comments.
		if strings.Contains(text, "\nvcl 4.0;") {
			return 0.9
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (VCL) Name() string {
	return heartbeat.LanguageVCL.StringChroma()
}
