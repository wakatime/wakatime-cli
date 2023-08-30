package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var objectiveJAnalyserImportRe = regexp.MustCompile(`(?m)^\s*@import\s+[<"]`)

// ObjectiveJ lexer.
type ObjectiveJ struct{}

// Lexer returns the lexer.
func (l ObjectiveJ) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"objective-j", "objectivej", "obj-j", "objj"},
			Filenames: []string{"*.j"},
			MimeTypes: []string{"text/x-objective-j"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// special directive found in most Objective-J files.
		if objectiveJAnalyserImportRe.MatchString(text) {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (ObjectiveJ) Name() string {
	return heartbeat.LanguageObjectiveJ.StringChroma()
}
