package language

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

// detectFSharpFromContents tries to detect the language from the file contents.
func detectFSharpFromContents(text string) (heartbeat.Language, float32, bool) {
	var weight float32

	if strings.Contains(text, "let ") && strings.Contains(text, "match ") && strings.Contains(text, " ->") {
		weight = 0.9
	}

	if strings.Contains(text, "// ") || strings.Contains(text, "(* ") && strings.Contains(text, " *)") {
		weight += 0.7
	}

	if weight > 1 {
		weight = 1
	}

	return heartbeat.LanguageUnknown, weight, weight > 0
}
