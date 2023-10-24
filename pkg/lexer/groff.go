package lexer

import (
	"regexp"
	"unicode"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoglobals
var groffAlphanumericRe = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

// Groff lexer.
type Groff struct{}

// Lexer returns the lexer.
func (l Groff) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	var (
		ok       bool
		rgxlexer *chroma.RegexLexer
	)

	if rgxlexer, ok = lexer.(*chroma.RegexLexer); !ok {
		return nil
	}

	rgxlexer.SetAnalyser(func(text string) float32 {
		if len(text) <= 1 {
			return 0
		}

		if text[:1] != "." {
			return 0
		}

		if len(text) <= 3 {
			return 0
		}

		if text[:3] == `.\"` {
			return 1.0
		}

		if len(text) <= 4 {
			return 0
		}

		if text[:4] == ".TH " {
			return 1.0
		}

		if groffAlphanumericRe.MatchString(text[1:3]) && unicode.IsSpace(rune(text[3])) {
			return 0.9
		}

		return 0
	})

	return rgxlexer
}

// Name returns the name of the lexer.
func (Groff) Name() string {
	return heartbeat.LanguageGroff.StringChroma()
}
