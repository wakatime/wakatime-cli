package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/shebang"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var perlAnalyserRe = regexp.MustCompile(`(?:my|our)\s+[$@%(]`)

// Perl lexer.
type Perl struct{}

// Lexer returns the lexer.
func (l Perl) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			if matched, _ := shebang.MatchString(text, "perl"); matched {
				return 1.0
			}

			var result float32

			if perlAnalyserRe.MatchString(text) {
				result += 0.9
			}

			if strings.Contains(text, ":=") {
				// := is not valid Perl, but it appears in unicon, so we should
				// become less confident if we think we found Perl with :=
				result /= 2
			}

			return result
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Perl) Name() string {
	return heartbeat.LanguagePerl.StringChroma()
}
