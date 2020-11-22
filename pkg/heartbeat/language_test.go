package heartbeat_test

import (
	"encoding/json"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func languageTests() map[string]heartbeat.Language {
	return map[string]heartbeat.Language{
		"AppleScript":         heartbeat.LanguageAppleScript,
		"ApacheConf":          heartbeat.LanguageApacheConf,
		"Assembly":            heartbeat.LanguageAssembly,
		"Awk":                 heartbeat.LanguageAwk,
		"Bash":                heartbeat.LanguageBash,
		"C":                   heartbeat.LanguageC,
		"C++":                 heartbeat.LanguageCPP,
		"C#":                  heartbeat.LanguageCSharp,
		"Clojure":             heartbeat.LanguageClojure,
		"CMake":               heartbeat.LanguageCMake,
		"CoffeeScript":        heartbeat.LanguageCoffeeScript,
		"Coldfusion":          heartbeat.LanguageColdfusionHTML,
		"Common Lisp":         heartbeat.LanguageCommonLisp,
		"Crontab":             heartbeat.LanguageCrontab,
		"Crystal":             heartbeat.LanguageCrystal,
		"CSS":                 heartbeat.LanguageCSS,
		"Dart":                heartbeat.LanguageDart,
		"Delphi":              heartbeat.LanguageDelphi,
		"Docker":              heartbeat.LanguageDocker,
		"Elixir":              heartbeat.LanguageElixir,
		"Elm":                 heartbeat.LanguageElm,
		"Emacs Lisp":          heartbeat.LanguageEmacsLisp,
		"Erlang":              heartbeat.LanguageErlang,
		"F#":                  heartbeat.LanguageFSharp,
		"Fortran":             heartbeat.LanguageFortran,
		"Go":                  heartbeat.LanguageGo,
		"Gosu":                heartbeat.LanguageGosu,
		"Groovy":              heartbeat.LanguageGroovy,
		"Haskell":             heartbeat.LanguageHaskell,
		"Haxe":                heartbeat.LanguageHaxe,
		"HTML":                heartbeat.LanguageHTML,
		"INI":                 heartbeat.LanguageINI,
		"Java":                heartbeat.LanguageJava,
		"JavaScript":          heartbeat.LanguageJavaScript,
		"JSON":                heartbeat.LanguageJSON,
		"JSX":                 heartbeat.LanguageJSX,
		"Kotlin":              heartbeat.LanguageKotlin,
		"Lasso":               heartbeat.LanguageLasso,
		"TeX":                 heartbeat.LanguageTex,
		"LESS":                heartbeat.LanguageLess,
		"liquid":              heartbeat.LanguageLiquid,
		"Lua":                 heartbeat.LanguageLua,
		"Mako":                heartbeat.LanguageMako,
		"Markdown":            heartbeat.LanguageMarkdown,
		"Marko":               heartbeat.LanguageMarko,
		"Matlab":              heartbeat.LanguageMatlab,
		"Modelica":            heartbeat.LanguageModelica,
		"Modula-2":            heartbeat.LanguageModula,
		"Mustache":            heartbeat.LanguageMustache,
		"NewLisp":             heartbeat.LanguageNewLisp,
		"Nix":                 heartbeat.LanguageNix,
		"Objective-C":         heartbeat.LanguageObjectiveC,
		"Objective-C++":       heartbeat.LanguageObjectiveCPP,
		"Objective-J":         heartbeat.LanguageObjectiveJ,
		"OCaml":               heartbeat.LanguageOCaml,
		"Pawn":                heartbeat.LanguagePawn,
		"Perl":                heartbeat.LanguagePerl,
		"PHP":                 heartbeat.LanguagePHP,
		"PostScript":          heartbeat.LanguagePostScript,
		"POVRay":              heartbeat.LanguagePOVRay,
		"PowerShell":          heartbeat.LanguagePowerShell,
		"Prolog":              heartbeat.LanguageProlog,
		"Protocol Buffer":     heartbeat.LanguageProtocolBuffer,
		"Pug":                 heartbeat.LanguagePug,
		"Puppet":              heartbeat.LanguagePuppet,
		"Python":              heartbeat.LanguagePython,
		"QML":                 heartbeat.LanguageQML,
		"R":                   heartbeat.LanguageR,
		"ReasonML":            heartbeat.LanguageReasonML,
		"reStructuredText":    heartbeat.LanguageReStructuredText,
		"RPMSpec":             heartbeat.LanguageRPMSpec,
		"Ruby":                heartbeat.LanguageRuby,
		"Rust":                heartbeat.LanguageRust,
		"Sass":                heartbeat.LanguageSass,
		"Scala":               heartbeat.LanguageScala,
		"Scheme":              heartbeat.LanguageScheme,
		"SCSS":                heartbeat.LanguageSCSS,
		"Sketch Drawing":      heartbeat.LanguageSketchDrawing,
		"Slim":                heartbeat.LanguageSlim,
		"Smali":               heartbeat.LanguageSmali,
		"Smalltalk":           heartbeat.LanguageSmalltalk,
		"SourcePawn":          heartbeat.LanguageSourcePawn,
		"SQL":                 heartbeat.LanguageSQL,
		"Sublime Text Config": heartbeat.LanguageSublimeTextConfig,
		"Svelte":              heartbeat.LanguageSvelte,
		"Swift":               heartbeat.LanguageSwift,
		"SWIG":                heartbeat.LanguageSWIG,
		"systemverilog":       heartbeat.Languagesystemverilog,
		"Text":                heartbeat.LanguageText,
		"Thrift":              heartbeat.LanguageThrift,
		"TOML":                heartbeat.LanguageTOML,
		"Twig":                heartbeat.LanguageTwig,
		"TypeScript":          heartbeat.LanguageTypeScript,
		"TypoScript":          heartbeat.LanguageTypoScript,
		"VB.net":              heartbeat.LanguageVB,
		"VCL":                 heartbeat.LanguageVCL,
		"Velocity":            heartbeat.LanguageVelocity,
		"VimL":                heartbeat.LanguageVimL,
		"Vue.js":              heartbeat.LanguageVueJS,
		"XAML":                heartbeat.LanguageXAML,
		"XML":                 heartbeat.LanguageXML,
		"XSLT":                heartbeat.LanguageXSLT,
		"YAML":                heartbeat.LanguageYAML,
		"Zig":                 heartbeat.LanguageZig,
	}
}

func TestParseLanguage(t *testing.T) {
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			parsed, ok := heartbeat.ParseLanguage(value)
			assert.True(t, ok)

			assert.Equal(t, language, parsed)
		})
	}

	t.Run("lower case", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("go")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageGo, parsed)
	})

	t.Run("hash", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("CSharp")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageCSharp, parsed)
	})

	t.Run("plus sign", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("CPP")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageCPP, parsed)
	})

	t.Run("leading space", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage(" Go")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageGo, parsed)
	})

	t.Run("trailing space", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("Go ")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageGo, parsed)
	})

	t.Run("missing hyphen", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("ObjectiveC")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageObjectiveC, parsed)
	})

	t.Run("missing space", func(t *testing.T) {
		parsed, ok := heartbeat.ParseLanguage("Sublime Text Config")
		assert.True(t, ok)

		assert.Equal(t, heartbeat.LanguageSublimeTextConfig, parsed)
	})
}

func TestParseLanguage_Unknown(t *testing.T) {
	parsed, ok := heartbeat.ParseLanguage("invalid")

	assert.False(t, ok)
	assert.Equal(t, heartbeat.LanguageUnknown, parsed)
}

func TestParseLanguageFromChroma(t *testing.T) {
	tests := map[string]heartbeat.Language{
		"AppleScript":      heartbeat.LanguageAppleScript,
		"ApacheConf":       heartbeat.LanguageApacheConf,
		"Awk":              heartbeat.LanguageAwk,
		"Bash":             heartbeat.LanguageBash,
		"C":                heartbeat.LanguageC,
		"C++":              heartbeat.LanguageCPP,
		"C#":               heartbeat.LanguageCSharp,
		"Clojure":          heartbeat.LanguageClojure,
		"CMake":            heartbeat.LanguageCMake,
		"CoffeeScript":     heartbeat.LanguageCoffeeScript,
		"Common Lisp":      heartbeat.LanguageCommonLisp,
		"Crystal":          heartbeat.LanguageCrystal,
		"CSS":              heartbeat.LanguageCSS,
		"Dart":             heartbeat.LanguageDart,
		"Docker":           heartbeat.LanguageDocker,
		"Elixir":           heartbeat.LanguageElixir,
		"Elm":              heartbeat.LanguageElm,
		"EmacsLisp":        heartbeat.LanguageEmacsLisp,
		"Erlang":           heartbeat.LanguageErlang,
		"FSharp":           heartbeat.LanguageFSharp,
		"Fortran":          heartbeat.LanguageFortran,
		"GAS":              heartbeat.LanguageAssembly,
		"Go":               heartbeat.LanguageGo,
		"Groovy":           heartbeat.LanguageGroovy,
		"Haskell":          heartbeat.LanguageHaskell,
		"Haxe":             heartbeat.LanguageHaxe,
		"HTML":             heartbeat.LanguageHTML,
		"INI":              heartbeat.LanguageINI,
		"Java":             heartbeat.LanguageJava,
		"JavaScript":       heartbeat.LanguageJavaScript,
		"JSON":             heartbeat.LanguageJSON,
		"Kotlin":           heartbeat.LanguageKotlin,
		"TeX":              heartbeat.LanguageTex,
		"Lua":              heartbeat.LanguageLua,
		"Mako":             heartbeat.LanguageMako,
		"markdown":         heartbeat.LanguageMarkdown,
		"Matlab":           heartbeat.LanguageMatlab,
		"Modula-2":         heartbeat.LanguageModula,
		"Nix":              heartbeat.LanguageNix,
		"Objective-C":      heartbeat.LanguageObjectiveC,
		"OCaml":            heartbeat.LanguageOCaml,
		"Perl":             heartbeat.LanguagePerl,
		"PHP":              heartbeat.LanguagePHP,
		"plaintext":        heartbeat.LanguageText,
		"PostScript":       heartbeat.LanguagePostScript,
		"POVRay":           heartbeat.LanguagePOVRay,
		"PowerShell":       heartbeat.LanguagePowerShell,
		"Prolog":           heartbeat.LanguageProlog,
		"Protocol Buffer":  heartbeat.LanguageProtocolBuffer,
		"Puppet":           heartbeat.LanguagePuppet,
		"Python":           heartbeat.LanguagePython,
		"R":                heartbeat.LanguageR,
		"ReasonML":         heartbeat.LanguageReasonML,
		"reStructuredText": heartbeat.LanguageReStructuredText,
		"Ruby":             heartbeat.LanguageRuby,
		"Rust":             heartbeat.LanguageRust,
		"Sass":             heartbeat.LanguageSass,
		"Scala":            heartbeat.LanguageScala,
		"Scheme":           heartbeat.LanguageScheme,
		"SCSS":             heartbeat.LanguageSCSS,
		"Smalltalk":        heartbeat.LanguageSmalltalk,
		"SQL":              heartbeat.LanguageSQL,
		"Swift":            heartbeat.LanguageSwift,
		"systemverilog":    heartbeat.Languagesystemverilog,
		"Thrift":           heartbeat.LanguageThrift,
		"TOML":             heartbeat.LanguageTOML,
		"Twig":             heartbeat.LanguageTwig,
		"TypeScript":       heartbeat.LanguageTypeScript,
		"TypoScript":       heartbeat.LanguageTypoScript,
		"VB.net":           heartbeat.LanguageVB,
		"VimL":             heartbeat.LanguageVimL,
		"XML":              heartbeat.LanguageXML,
		"YAML":             heartbeat.LanguageYAML,
		"Zig":              heartbeat.LanguageZig,
	}

	for lexerName, language := range tests {
		t.Run(lexerName, func(t *testing.T) {
			parsed, ok := heartbeat.ParseLanguageFromChroma(lexerName)

			assert.True(t, ok)
			assert.Equal(t, language, parsed)
		})
	}
}

func TestParseLanguageFromChroma_Unknown(t *testing.T) {
	parsed, ok := heartbeat.ParseLanguageFromChroma("invalid")

	assert.False(t, ok)
	assert.Equal(t, heartbeat.LanguageUnknown, parsed)
}

func TestLanguage_MarshalJSON(t *testing.T) {
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			data, err := json.Marshal(language)
			require.NoError(t, err)

			assert.JSONEq(t, `"`+value+`"`, string(data))
		})
	}
}

func TestLanguage_MarshalJSON_UnknownLanguage(t *testing.T) {
	data, err := json.Marshal(heartbeat.LanguageUnknown)
	require.NoError(t, err)

	assert.JSONEq(t, `null`, string(data))
}

func TestLanguage_UnmarshalJSON(t *testing.T) {
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			var l heartbeat.Language
			require.NoError(t, json.Unmarshal([]byte(`"`+value+`"`), &l))

			assert.Equal(t, language, l)
		})
	}
}

func TestLanguage_String(t *testing.T) {
	for value, language := range languageTests() {
		t.Run(value, func(t *testing.T) {
			assert.Equal(t, value, language.String())
		})
	}
}

func TestLanguage_String_UnknownLanguage(t *testing.T) {
	assert.Equal(t, "Unknown", heartbeat.LanguageUnknown.String())
}
