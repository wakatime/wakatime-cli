package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageVCL.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"vcl"},
			Filenames: []string{"*.vcl"},
			MimeTypes: []string{"text/x-vclsrc"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// If the very first line is 'vcl 4.0;' it's pretty much guaranteed
		// that this is VCL
		if strings.HasPrefix(text, "vcl 4.0;") {
			return 1.0
		}

		if len(text) > 1000 {
			text = text[:1000]
		}

		// Skip over comments and blank lines
		// This is accurate enough that returning 0.9 is reasonable.
		// Almost no VCL files start without some comments.
		if strings.Contains(text, "\nvcl 4.0;") {
			return 0.9
		}

		return 0
	}))
}
