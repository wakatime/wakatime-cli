package language

import (
	"regexp"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

var modelineRegex = regexp.MustCompile(`(?m)(?:vi|vim|ex)(?:[<=>]?\d*)?:.*(?:ft|filetype|syn|syntax)=([^:\s]+)`)

// detectVimModeline tries to detect the language from the vim modeline.
func detectVimModeline(text string) (heartbeat.Language, float32, bool) {
	matches := modelineRegex.FindStringSubmatch(text)

	if matches == nil || len(matches) != 2 {
		return heartbeat.LanguageUnknown, 0, false
	}

	lang, ok := parseVim(matches[1])
	if !ok {
		return heartbeat.LanguageUnknown, 0, false
	}

	lexer := lexers.Get(lang.StringChroma())
	if lexer == nil {
		return heartbeat.LanguageUnknown, 0, false
	}

	analyser, ok := lexer.(chroma.Analyser)
	if !ok {
		return heartbeat.LanguageUnknown, 0, false
	}

	return lang, analyser.AnalyseText(text), true
}

// nolint:gocyclo
// parseVim parses the language from a vim plugin specific string.
func parseVim(language string) (heartbeat.Language, bool) {
	switch strings.ToLower(language) {
	case strings.ToLower("a65"):
		return heartbeat.ParseLanguage("assembly")
	case strings.ToLower("asm"):
		return heartbeat.ParseLanguage("assembly")
	case strings.ToLower("asm68k"):
		return heartbeat.ParseLanguage("assembly")
	case strings.ToLower("asmh8300"):
		return heartbeat.ParseLanguage("assembly")
	case strings.ToLower("basic"):
		return heartbeat.ParseLanguage("basic")
	case strings.ToLower("c"):
		return heartbeat.ParseLanguage("c")
	case strings.ToLower("cpp"):
		return heartbeat.ParseLanguage("cpp")
	case strings.ToLower("crontab"):
		return heartbeat.ParseLanguage("crontab")
	case strings.ToLower("cs"):
		return heartbeat.ParseLanguage("csharp")
	case strings.ToLower("haml"):
		return heartbeat.ParseLanguage("haml")
	case strings.ToLower("haskell"):
		return heartbeat.ParseLanguage("haskell")
	case strings.ToLower("html"):
		return heartbeat.ParseLanguage("html")
	case strings.ToLower("htmlcheetah"):
		return heartbeat.ParseLanguage("html")
	case strings.ToLower("htmldjango"):
		return heartbeat.ParseLanguage("html")
	case strings.ToLower("htmlm4"):
		return heartbeat.ParseLanguage("html")
	case strings.ToLower("java"):
		return heartbeat.ParseLanguage("java")
	case strings.ToLower("javascript"):
		return heartbeat.ParseLanguage("javascript")
	case strings.ToLower("lhaskell"):
		return heartbeat.ParseLanguage("haskell")
	case strings.ToLower("markdown"):
		return heartbeat.ParseLanguage("markdown")
	case strings.ToLower("objc"):
		return heartbeat.ParseLanguage("objectivec")
	case strings.ToLower("objcpp"):
		return heartbeat.ParseLanguage("objectivecpp")
	case strings.ToLower("ocaml"):
		return heartbeat.ParseLanguage("ocaml")
	case strings.ToLower("perl"):
		return heartbeat.ParseLanguage("perl")
	case strings.ToLower("perl6"):
		return heartbeat.ParseLanguage("perl")
	case strings.ToLower("php"):
		return heartbeat.ParseLanguage("php")
	case strings.ToLower("phtml"):
		return heartbeat.ParseLanguage("php")
	case strings.ToLower("prolog"):
		return heartbeat.ParseLanguage("prolog")
	case strings.ToLower("python"):
		return heartbeat.ParseLanguage("python")
	case strings.ToLower("r"):
		return heartbeat.ParseLanguage("r")
	case strings.ToLower("ruby"):
		return heartbeat.ParseLanguage("ruby")
	case strings.ToLower("sass"):
		return heartbeat.ParseLanguage("sass")
	case strings.ToLower("scheme"):
		return heartbeat.ParseLanguage("scheme")
	case strings.ToLower("scss"):
		return heartbeat.ParseLanguage("scss")
	case strings.ToLower("skill"):
		return heartbeat.ParseLanguage("skill")
	case strings.ToLower("vb"):
		return heartbeat.ParseLanguage("vbnet")
	case strings.ToLower("vim"):
		return heartbeat.ParseLanguage("viml")
	case strings.ToLower("xhtml"):
		return heartbeat.ParseLanguage("html")
	case strings.ToLower("xml"):
		return heartbeat.ParseLanguage("xml")
	case strings.ToLower("yaml"):
		return heartbeat.ParseLanguage("yaml")
	default:
		return heartbeat.LanguageUnknown, false
	}
}
