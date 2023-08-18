package lexer

import (
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/shebang"

	"github.com/alecthomas/chroma/v2"
)

// NumPy lexer.
type NumPy struct{}

// Lexer returns the lexer.
func (l NumPy) Lexer() chroma.Lexer {
	lexer := chroma.MustNewLexer(
		&chroma.Config{
			Name:    l.Name(),
			Aliases: []string{"numpy"},
		},
		func() chroma.Rules {
			return chroma.Rules{
				"root": {},
			}
		},
	)

	lexer.SetAnalyser(func(text string) float32 {
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
	})

	return lexer
}

// Name returns the name of the lexer.
func (NumPy) Name() string {
	return heartbeat.LanguageNumPy.StringChroma()
}
