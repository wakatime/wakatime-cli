package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	rebolAnalyserHeaderRe              = regexp.MustCompile(`^\s*REBOL\s*\[`)
	rebolAnalyserHeaderPrecedingTextRe = regexp.MustCompile(`\s*REBOL\s*\[`)
)

// REBOL lexer.
type REBOL struct{}

// Lexer returns the lexer.
func (l REBOL) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"rebol"},
			Filenames: []string{"*.r", "*.r3", "*.reb"},
			MimeTypes: []string{"text/x-rebol"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		// Check if code contains REBOL header, then it's probably not R code
		if rebolAnalyserHeaderRe.MatchString(text) {
			return 1.0
		}

		if rebolAnalyserHeaderPrecedingTextRe.MatchString(text) {
			return 0.5
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (REBOL) Name() string {
	return heartbeat.LanguageREBOL.StringChroma()
}
