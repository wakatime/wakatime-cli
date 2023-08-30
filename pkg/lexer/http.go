package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// HTTP lexer.
type HTTP struct{}

// Lexer returns the lexer.
func (HTTP) Lexer() chroma.Lexer {
	return lexers.HTTP.SetAnalyser(func(text string) float32 {
		if strings.HasPrefix(text, "GET") ||
			strings.HasPrefix(text, "POST") ||
			strings.HasPrefix(text, "PUT") ||
			strings.HasPrefix(text, "DELETE") ||
			strings.HasPrefix(text, "HEAD") ||
			strings.HasPrefix(text, "OPTIONS") ||
			strings.HasPrefix(text, "TRACE") ||
			strings.HasPrefix(text, "PATCH") {
			return 1.0
		}

		return 0
	})
}

// Name returns the name of the lexer.
func (HTTP) Name() string {
	return heartbeat.LanguageHTTP.StringChroma()
}
