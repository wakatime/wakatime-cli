package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageRuby.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		return
	}

	cfg := lexer.Config()
	if cfg == nil {
		return
	}

	cfg.Filenames = append(cfg.Filenames, ".ruby-version")
}
