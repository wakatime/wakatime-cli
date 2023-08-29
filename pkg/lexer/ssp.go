package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/xml"

	"github.com/alecthomas/chroma/v2"
)

var sspAnalyserRe = regexp.MustCompile(`val \w+\s*:`)

// SSP lexer. Lexer for Scalate Server Pages.
type SSP struct{}

// Lexer returns the lexer.
func (l SSP) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"ssp"},
			Filenames: []string{"*.ssp"},
			MimeTypes: []string{"application/x-ssp"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		var result float64

		if sspAnalyserRe.MatchString(text) {
			result += 0.6
		}

		if xml.MatchString(text) {
			result += 0.2
		}

		if strings.Contains(text, "<%") && strings.Contains(text, "%>") {
			result += 0.1
		}

		return float32(result)
	})

	return lexer
}

// Name returns the name of the lexer.
func (SSP) Name() string {
	return heartbeat.LanguageSSP.StringChroma()
}
