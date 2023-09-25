package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	matlabAnalyserCommentRe   = regexp.MustCompile(`^\s*%`)
	matlabAnalyserSystemCMDRe = regexp.MustCompile(`^!\w+`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageMatlab.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

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
}
