package language

import (
	"fmt"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectVimModeline(t *testing.T) {
	tests := map[string]struct {
		Text     string
		Language heartbeat.Language
		Error    error
	}{
		"ft": {
			Text:     `/* vim: ft=python tw=60 ts=2: */`,
			Language: heartbeat.LanguagePython,
		},
		"filetype": {
			Text:     "/* vim: filetype=python tw=60 ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"syn": {
			Text:     "/* vim: syn=python tw=60 ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"syntax": {
			Text:     "/* vim: syntax=python tw=60 ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"different order": {
			Text:     "/* vim: tw=60 ft=python ts=2: */",
			Language: heartbeat.LanguagePython,
		},
		"multiline": {
			Text: `
			/* vim: tw=60 ft=python ts=2: */
			`,
			Language: heartbeat.LanguagePython,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			lang, weight, ok := detectVimModeline(test.Text)
			require.True(t, ok)

			assert.Equal(t, float32(0), weight)
			assert.Equal(t, test.Language, lang, fmt.Sprintf("got: %q, want: %q", lang, test.Language))
		})
	}
}

func TestParseVim(t *testing.T) {
	tests := map[string]heartbeat.Language{
		"a65":         heartbeat.LanguageAssembly,
		"asm":         heartbeat.LanguageAssembly,
		"asm68k":      heartbeat.LanguageAssembly,
		"asmh8300":    heartbeat.LanguageAssembly,
		"basic":       heartbeat.LanguageBasic,
		"c":           heartbeat.LanguageC,
		"cpp":         heartbeat.LanguageCPP,
		"crontab":     heartbeat.LanguageCrontab,
		"cs":          heartbeat.LanguageCSharp,
		"haml":        heartbeat.LanguageHaml,
		"haskell":     heartbeat.LanguageHaskell,
		"html":        heartbeat.LanguageHTML,
		"htmlcheetah": heartbeat.LanguageHTML,
		"htmldjango":  heartbeat.LanguageHTML,
		"htmlm4":      heartbeat.LanguageHTML,
		"java":        heartbeat.LanguageJava,
		"javascript":  heartbeat.LanguageJavaScript,
		"lhaskell":    heartbeat.LanguageHaskell,
		"markdown":    heartbeat.LanguageMarkdown,
		"objc":        heartbeat.LanguageObjectiveC,
		"objcpp":      heartbeat.LanguageObjectiveCPP,
		"ocaml":       heartbeat.LanguageOCaml,
		"perl":        heartbeat.LanguagePerl,
		"perl6":       heartbeat.LanguagePerl,
		"php":         heartbeat.LanguagePHP,
		"phtml":       heartbeat.LanguagePHP,
		"prolog":      heartbeat.LanguageProlog,
		"python":      heartbeat.LanguagePython,
		"r":           heartbeat.LanguageR,
		"ruby":        heartbeat.LanguageRuby,
		"sass":        heartbeat.LanguageSass,
		"scheme":      heartbeat.LanguageScheme,
		"scss":        heartbeat.LanguageSCSS,
		"skill":       heartbeat.LanguageSKILL,
		"vb":          heartbeat.LanguageVBNet,
		"vim":         heartbeat.LanguageVimL,
		"xhtml":       heartbeat.LanguageHTML,
		"xml":         heartbeat.LanguageXML,
		"yaml":        heartbeat.LanguageYAML,
		// upper case should also be accepted
		"YAML": heartbeat.LanguageYAML,
	}

	for name, lang := range tests {
		t.Run(name, func(t *testing.T) {
			parsed, ok := parseVim(name)
			require.True(t, ok)

			assert.Equal(t, lang, parsed, fmt.Sprintf("got: %q, want: %q", parsed, lang))
		})
	}
}
