package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var sourcesListAnalyserRe = regexp.MustCompile(`(?m)^\s*(deb|deb-src) `)

// SourcesList lexer. Lexer that highlights debian sources.list files.
type SourcesList struct{}

// Lexer returns the lexer.
func (l SourcesList) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"sourceslist", "sources.list", "debsources"},
			Filenames: []string{"sources.list"},
			MimeTypes: []string{"application/x-debian-sourceslist"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if sourcesListAnalyserRe.MatchString(text) {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (SourcesList) Name() string {
	return heartbeat.LanguageSourcesList.StringChroma()
}
