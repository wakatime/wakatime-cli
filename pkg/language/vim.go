package language

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
)

var modelineRegex = regexp.MustCompile(`(?m)(?:vi|vim|ex)(?:[<=>]?\d*)?:.*(?:ft|filetype|syn|syntax)=([^:\s]+)`)

// detectVimModeline tries to detect the language from the vim modeline.
func detectVimModeline(text string) (heartbeat.Language, float32, bool) {
	matches := modelineRegex.FindStringSubmatch(text)

	if matches == nil || len(matches) != 2 {
		return heartbeat.LanguageUnknown, 0, false
	}

	lang, ok := Parse(matches[1], "vim")
	if !ok {
		return heartbeat.LanguageUnknown, 0, false
	}

	lexer := lexers.Get(lang.StringChroma())
	if lexer == nil {
		return heartbeat.LanguageUnknown, 0, false
	}

	analyser, ok := lexer.(chroma.Analyser)
	if !ok {
		return heartbeat.LanguageUnknown, 0, false
	}

	return lang, analyser.AnalyseText(text), true
}
