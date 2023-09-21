package lexer

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

var (
	// Have to be careful not to accidentally match JavaDoc/Doxygen syntax here,
	// since that's quite common in ordinary C/C++ files.  It's OK to match
	// JavaDoc/Doxygen keywords that only apply to Objective-C, mind.
	//
	// The upshot of this is that we CANNOT match @class or @interface.
	objectiveCAnalyserKeywordsRe = regexp.MustCompile(`@(?:end|implementation|protocol)`)
	// Matches [ <ws>? identifier <ws> ( identifier <ws>? ] |  identifier? : )
	// (note the identifier is *optional* when there is a ':'!)
	objectiveCAnalyserMessageRe  = regexp.MustCompile(`\[\s*[a-zA-Z_]\w*\s+(?:[a-zA-Z_]\w*\s*\]|(?:[a-zA-Z_]\w*)?:)`)
	objectiveCAnalyserNSNumberRe = regexp.MustCompile(`@[0-9]+`)
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageObjectiveC.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	lexer.SetAnalyser(func(text string) float32 {
		if objectiveCAnalyserKeywordsRe.MatchString(text) {
			return 1.0
		}

		if strings.Contains(text, `@"`) {
			return 0.8
		}

		if objectiveCAnalyserNSNumberRe.MatchString(text) {
			return 0.7
		}

		if objectiveCAnalyserMessageRe.MatchString(text) {
			return 0.8
		}

		return 0
	})
}
