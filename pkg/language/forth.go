package language

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

var forthFuncTest = regexp.MustCompile(`:[^\n\r]+;[\n\r]`)

// detectForthFromContents tries to detect the language from the file contents.
func detectForthFromContents(text string) (heartbeat.Language, float32, bool) {
	var weight float32

	if forthFuncTest.MatchString(text) {
		weight = 0.9
	}

	if strings.Contains(text, "\\ ") {
		weight += 0.5
	}

	if strings.Contains(text, "( ") {
		weight += 0.2
	}

	if weight > 1 {
		weight = 1
	}

	return heartbeat.LanguageUnknown, weight, weight > 0
}
