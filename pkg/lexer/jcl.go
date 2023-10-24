package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var jclAnalyserJobHeaderRe = regexp.MustCompile(`(?i)^//[a-z#$@][a-z0-9#$@]{0,7}\s+job(\s+.*)?$`)

// JCL lexer.
type JCL struct{}

// Lexer returns the lexer.
func (l JCL) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"jcl"},
			Filenames: []string{"*.jcl"},
			MimeTypes: []string{"text/x-jcl"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// Recognize JCL job by header.
		lines := strings.Split(text, "\n")
		if len(lines) == 0 {
			return 0
		}

		if jclAnalyserJobHeaderRe.MatchString(lines[0]) {
			return 1.0
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (JCL) Name() string {
	return heartbeat.LanguageJCL.StringChroma()
}
