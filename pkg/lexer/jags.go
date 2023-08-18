package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	jagsAnalyserModelRe = regexp.MustCompile(`(?m)^\s*model\s*\{`)
	jagsAnalyserDataRe  = regexp.MustCompile(`(?m)^\s*data\s*\{`)
	jagsAnalyserVarRe   = regexp.MustCompile(`(?m)^\s*var`)
)

// JAGS lexer.
type JAGS struct{}

// Lexer returns the lexer.
func (l JAGS) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"jags"},
			Filenames: []string{"*.jag", "*.bug"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if jagsAnalyserModelRe.MatchString(text) {
			if jagsAnalyserDataRe.MatchString(text) {
				return 0.9
			}

			if jagsAnalyserVarRe.MatchString(text) {
				return 0.9
			}

			return 0.3
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (JAGS) Name() string {
	return heartbeat.LanguageJAGS.StringChroma()
}
