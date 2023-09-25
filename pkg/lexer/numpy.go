package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/shebang"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// nolint:gochecknoinits
func init() {
	language := heartbeat.LanguageNumPy.StringChroma()
	lexer := lexers.Get(language)

	if lexer != nil {
		log.Debugf("lexer %q already registered", language)
		return
	}

	_ = lexers.Register(chroma.MustNewLexer(
		&chroma.Config{
			Name:    language,
			Aliases: []string{"numpy"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	).SetAnalyser(func(text string) float32 {
		hasPythonShebang, _ := shebang.MatchString(text, `pythonw?(3(\.\d)?)?`)
		containsNumpyImport := strings.Contains(text, "import numpy")
		containsFromNumpyImport := strings.Contains(text, "from numpy import")

		var containsImport bool

		if len(text) > 1000 {
			containsImport = strings.Contains(text[:1000], "import ")
		} else {
			containsImport = strings.Contains(text, "import ")
		}

		if (hasPythonShebang || containsImport) && (containsNumpyImport || containsFromNumpyImport) {
			return 1.0
		}

		return 0
	}))
}
