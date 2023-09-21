package lexer

import (
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageObjectiveCPP.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"objective-c++", "objectivec++", "obj-c++", "objc++"},
			Filenames: []string{"*.mm", "*.hh"},
			MimeTypes: []string{"text/x-objective-c++"},
			Priority:  0.05,
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	))
}
