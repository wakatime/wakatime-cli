package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// VBScript lexer.
type VBScript struct{}

// Lexer returns the lexer.
func (l VBScript) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"vbscript"},
			Filenames: []string{"*.vbs", "*.VBS"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (VBScript) Name() string {
	return heartbeat.LanguageVBScript.StringChroma()
}
