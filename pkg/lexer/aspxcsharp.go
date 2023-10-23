package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
)

var (
	csharpAspxAnalyzerPageLanguageRe   = regexp.MustCompile(`(?i)Page\s*Language="C#"`)
	csharpAspxAnalyzerScriptLanguageRe = regexp.MustCompile(`(?i)script[^>]+language=["\']C#`)
)

// AspxCSharp lexer.
type AspxCSharp struct{}

// Lexer returns the lexer.
func (l AspxCSharp) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:      l.Name(),
			Aliases:   []string{"aspx-cs"},
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
		if csharpAspxAnalyzerPageLanguageRe.MatchString(text) {
			return 0.2
		}

		if csharpAspxAnalyzerScriptLanguageRe.MatchString(text) {
			return 0.15
		}

		return 0
	})

	return lexer
}

// Name returns the name of the lexer.
func (AspxCSharp) Name() string {
	return heartbeat.LanguageAspxCSharp.StringChroma()
}
