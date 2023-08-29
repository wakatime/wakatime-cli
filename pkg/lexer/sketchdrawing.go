package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// SketchDrawing lexer.
type SketchDrawing struct{}

// Lexer returns the lexer.
func (l SketchDrawing) Lexer() chroma.Lexer {
	return chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"sketch"},
			Filenames: []string{"*.sketch"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)
}

// Name returns the name of the lexer.
func (SketchDrawing) Name() string {
	return heartbeat.LanguageSketchDrawing.StringChroma()
}
