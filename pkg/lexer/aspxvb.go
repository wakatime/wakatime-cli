package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

// nolint:gochecknoglobals
var (
	vbAspxAnalyzerPageLanguageRe   = regexp.MustCompile(`(?i)Page\s*Language="Vb"`)
	vbAspxAnalyzerScriptLanguageRe = regexp.MustCompile(`(?i)script[^>]+language=["\']vb`)
)

// AspxVBNet lexer.
type AspxVBNet struct{}

// Lexer returns the lexer.
func (l AspxVBNet) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"aspx-vb"},
			Filenames: []string{"*.aspx", "*.asax", "*.ascx", "*.ashx", "*.asmx", "*.axd"},
			MimeTypes: []string{},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
		if vbAspxAnalyzerPageLanguageRe.MatchString(text) {
			return 0.2
		}

		if vbAspxAnalyzerScriptLanguageRe.MatchString(text) {
			return 0.15
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (AspxVBNet) Name() string {
	return heartbeat.LanguageAspxVBNet.StringChroma()
}
