package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	matlabAnalyserCommentRe   = regexp.MustCompile(`^\s*%`)
	matlabAnalyserSystemCMDRe = regexp.MustCompile(`^!\w+`)
)

// Matlab lexer.
type Matlab struct{}

// Lexer returns the lexer.
func (l Matlab) Lexer() chroma.Lexer {
	lexer := lexers.Get(l.Name())
	if lexer == nil {
		return nil
	}

	if lexer, ok := lexer.(*chroma.RegexLexer); ok {
		lexer.SetAnalyser(func(text string) float32 {
			lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

			var firstNonComment string
			for _, line := range lines {
				if !matlabAnalyserCommentRe.MatchString(line) {
					firstNonComment = strings.TrimSpace(line)
					break
				}
			}

			// function declaration
			if strings.HasPrefix(firstNonComment, "function") && !strings.Contains(firstNonComment, "{") {
				return 1.0
			}

			// comment
			for _, line := range lines {
				if matlabAnalyserCommentRe.MatchString(line) {
					return 0.2
				}
			}

			// system cmd
			for _, line := range lines {
				if matlabAnalyserSystemCMDRe.MatchString(line) {
					return 0.2
				}
			}

			return 0
		})

		return lexer
	}

	return nil
}

// Name returns the name of the lexer.
func (Matlab) Name() string {
	return heartbeat.LanguageMatlab.StringChroma()
}
