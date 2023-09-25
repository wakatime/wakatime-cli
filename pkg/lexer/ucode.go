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
	language := heartbeat.LanguageUcode.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:      language,
			Aliases:   []string{"ucode"},
			Filenames: []string{"*.u", "*.u1", "*.u2"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		// endsuspend and endrepeat are unique to this language, and
		// \self, /self doesn't seem to get used anywhere else either.
		var result float32

		if strings.Contains(text, "endsuspend") {
			result += 0.1
		}

		if strings.Contains(text, "endrepeat") {
			result += 0.1
		}

		if strings.Contains(text, ":=") {
			result += 0.01
		}

		if strings.Contains(text, "procedure") && strings.Contains(text, "end") {
			result += 0.01
		}

		// This seems quite unique to unicon -- doesn't appear in any other
		// example source we have (A quick search reveals that \SELF appears in
		// Perl/Raku code)
		if strings.Contains(text, `\self`) && strings.Contains(text, "/self") {
			result += 0.5
		}

		return result
	}))
}
