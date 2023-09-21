package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguagePythonTraceback.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"pytb", "py3tb"},
			Filenames: []string{"*.pytb", "*.py3tb"},
			MimeTypes: []string{"text/x-python-traceback", "text/x-python3-traceback"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	))
}
