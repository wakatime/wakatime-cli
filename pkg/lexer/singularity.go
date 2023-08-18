package lexer

import (
	"regexp"

	"github.com/alecthomas/chroma/v2"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

var (
	singularityAnalyserHeaderRe  = regexp.MustCompile(`(?i)\b(?:osversion|includecmd|mirrorurl)\b`)
	singularityAnalyserSectionRe = regexp.MustCompile(
		`%(?:pre|post|setup|environment|help|labels|test|runscript|files|startscript)\b`)
)

// Singularity lexer.
type Singularity struct{}

// Lexer returns the lexer.
func (l Singularity) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"singularity"},
			Filenames: []string{"*.def", "Singularity"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// This is a quite simple script file, but there are a few keywords
		// which seem unique to this language.
		var result float32

		if singularityAnalyserHeaderRe.MatchString(text) {
			result += 0.5
		}

		if singularityAnalyserSectionRe.MatchString(text) {
			result += 0.49
		}

		return result
	})

	return lexer
}

// Name returns the name of the lexer.
func (Singularity) Name() string {
	return heartbeat.LanguageSingularity.StringChroma()
}
