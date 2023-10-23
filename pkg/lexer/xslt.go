package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/xml"

	"github.com/alecthomas/chroma/v2"
)

// XSLT lexer.
type XSLT struct{}

// Lexer returns the lexer.
func (XSLT) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:    "XSLT",
			Aliases: []string{"xslt"},
			// xpl is XProc
			Filenames: []string{"*.xsl", "*.xslt", "*.xpl"},
			MimeTypes: []string{"application/xsl+xml", "application/xslt+xml"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if xml.MatchString(text) && strings.Contains(text, "<xsl") {
			return 0.8
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (XSLT) Name() string {
	return heartbeat.LanguageXSLT.StringChroma()
}
