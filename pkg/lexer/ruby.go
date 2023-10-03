package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageRuby.StringChroma()
	lexer := lexers.Get(language)

	if lexer == nil {
		log.Debugf("lexer %q not found", language)
		return
	}

	cfg := lexer.Config()
	if cfg == nil {
		log.Debugf("lexer %q config not found", language)
		return
	}

	cfg.Filenames = append(cfg.Filenames, ".ruby-version")
}
