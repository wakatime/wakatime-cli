package lexer

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	csharpAspxAnalyzerPageLanguageRe   = regexp.MustCompile(`(?i)Page\s*Language="C#"`)
	csharpAspxAnalyzerScriptLanguageRe = regexp.MustCompile(`(?i)script[^>]+language=["\']C#`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageAspxCSharp.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"aspx-cs"},
			Filenames: []string{"*.aspx", "*.asax", "*.ascx", "*.ashx", "*.asmx", "*.axd"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		if csharpAspxAnalyzerPageLanguageRe.MatchString(text) {
			return 0.2
		}

		if csharpAspxAnalyzerScriptLanguageRe.MatchString(text) {
			return 0.15
		}

		return 0
	}))
}
