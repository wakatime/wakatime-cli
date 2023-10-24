package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// nolint:gochecknoglobals
var ca65AnalyserCommentRe = regexp.MustCompile(`(?m)^\s*;`)

// Ca65Assembler lexer.
type Ca65Assembler struct{}

// Lexer returns the lexer.
func (l Ca65Assembler) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ca65"},
			Filenames: []string{"*.s"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// comments in GAS start with "#".
		if ca65AnalyserCommentRe.MatchString(text) {
			return 0.9
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (Ca65Assembler) Name() string {
	return heartbeat.LanguageCa65Assembler.StringChroma()
}
