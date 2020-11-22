package heartbeat

import (
	"fmt"
	"strings"
)

// Language represents a programming language.
type Language int

const (
	// LanguageUnknown represents the Unknown programming language.
	LanguageUnknown Language = iota
	// LanguageAppleScript represents the AppleScript programming language.
	LanguageAppleScript
	// LanguageApacheConf represents the ApacheConf programming language.
	LanguageApacheConf
	// LanguageAssembly represents the Assembly programming language.
	LanguageAssembly
	// LanguageAwk represents the Awk programming language.
	LanguageAwk
	// LanguageBash represents the Bash programming language.
	LanguageBash
	// LanguageBasic represents the Basic programming language.
	LanguageBasic
	// LanguageC represents the C programming language.
	LanguageC
	// LanguageCPP represents the CPP programming language.
	LanguageCPP
	// LanguageCSharp represents the CSharp programming language.
	LanguageCSharp
	// LanguageClojure represents the Clojure programming language.
	LanguageClojure
	// LanguageCMake represents the CMake programming language.
	LanguageCMake
	// LanguageCoffeeScript represents the CoffeeScript programming language.
	LanguageCoffeeScript
	// LanguageColdfusionHTML represents the ColdfusionHTML programming language.
	LanguageColdfusionHTML
	// LanguageCommonLisp represents the CommonLisp programming language.
	LanguageCommonLisp
	// LanguageCrontab represents the Crontab programming language.
	LanguageCrontab
	// LanguageCrystal represents the Crystal programming language.
	LanguageCrystal
	// LanguageCSS represents the CSS programming language.
	LanguageCSS
	// LanguageDart represents the Dart programming language.
	LanguageDart
	// LanguageDelphi represents the Delphi programming language.
	LanguageDelphi
	// LanguageDocker represents the Docker programming language.
	LanguageDocker
	// LanguageElixir represents the Elixir programming language.
	LanguageElixir
	// LanguageElm represents the Elm programming language.
	LanguageElm
	// LanguageEmacsLisp represents the EmacsLisp programming language.
	LanguageEmacsLisp
	// LanguageErlang represents the Erlang programming language.
	LanguageErlang
	// LanguageFSharp represents the FSharp programming language.
	LanguageFSharp
	// LanguageFortran represents the Fortran programming language.
	LanguageFortran
	// LanguageGo represents the Go programming language.
	LanguageGo
	// LanguageGosu represents the Gosu programming language.
	LanguageGosu
	// LanguageGroovy represents the Groovy programming language.
	LanguageGroovy
	// LanguageHAML represents the HAML programming language.
	LanguageHAML
	// LanguageHaskell represents the Haskell programming language.
	LanguageHaskell
	// LanguageHaxe represents the Haxe programming language.
	LanguageHaxe
	// LanguageHTML represents the HTML programming language.
	LanguageHTML
	// LanguageINI represents the INI programming language.
	LanguageINI
	// LanguageJava represents the Java programming language.
	LanguageJava
	// LanguageJavaScript represents the JavaScript programming language.
	LanguageJavaScript
	// LanguageJSON represents the JSON programming language.
	LanguageJSON
	// LanguageJSX represents the JSX programming language.
	LanguageJSX
	// LanguageKotlin represents the Kotlin programming language.
	LanguageKotlin
	// LanguageLasso represents the Lasso programming language.
	LanguageLasso
	// LanguageTex represents the Tex programming language.
	LanguageTex
	// LanguageLess represents the Less programming language.
	LanguageLess
	// LanguageLiquid represents the Liquid programming language.
	LanguageLiquid
	// LanguageLua represents the Lua programming language.
	LanguageLua
	// LanguageMakefile represents the Makefile programming language.
	LanguageMakefile
	// LanguageMako represents the Mako programming language.
	LanguageMako
	// LanguageMarkdown represents the Markdown programming language.
	LanguageMarkdown
	// LanguageMarko represents the Marko programming language.
	LanguageMarko
	// LanguageMatlab represents the Matlab programming language.
	LanguageMatlab
	// LanguageModelica represents the Modelica programming language.
	LanguageModelica
	// LanguageModula represents the Modula programming language.
	LanguageModula
	// LanguageMustache represents the Mustache programming language.
	LanguageMustache
	// LanguageNewLisp represents the NewLisp programming language.
	LanguageNewLisp
	// LanguageNix represents the Nix programming language.
	LanguageNix
	// LanguageObjectiveC represents the ObjectiveC programming language.
	LanguageObjectiveC
	// LanguageObjectiveCPP represents the ObjectiveC++ programming language.
	LanguageObjectiveCPP
	// LanguageObjectiveJ represents the ObjectiveJ programming language.
	LanguageObjectiveJ
	// LanguageOCaml represents the OCaml programming language.
	LanguageOCaml
	// LanguagePawn represents the Pawn programming language.
	LanguagePawn
	// LanguagePerl represents the Perl programming language.
	LanguagePerl
	// LanguagePHP represents the PHP programming language.
	LanguagePHP
	// LanguagePostScript represents the PostScript programming language.
	LanguagePostScript
	// LanguagePOVRay represents the POVRay programming language.
	LanguagePOVRay
	// LanguagePowerShell represents the PowerShell programming language.
	LanguagePowerShell
	// LanguageProlog represents the Prolog programming language.
	LanguageProlog
	// LanguageProtocolBuffer represents the ProtocolBuffer programming language.
	LanguageProtocolBuffer
	// LanguagePug represents the Pug programming language.
	LanguagePug
	// LanguagePuppet represents the Puppet programming language.
	LanguagePuppet
	// LanguagePython represents the Python programming language.
	LanguagePython
	// LanguageQML represents the QML programming language.
	LanguageQML
	// LanguageR represents the R programming language.
	LanguageR
	// LanguageReasonML represents the ReasonML programming language.
	LanguageReasonML
	// LanguageReStructuredText represents the ReStructuredText programming language.
	LanguageReStructuredText
	// LanguageRPMSpec represents the RPMSpec programming language.
	LanguageRPMSpec
	// LanguageRuby represents the Ruby programming language.
	LanguageRuby
	// LanguageRust represents the Rust programming language.
	LanguageRust
	// LanguageSass represents the Sass programming language.
	LanguageSass
	// LanguageScala represents the Scala programming language.
	LanguageScala
	// LanguageScheme represents the Scheme programming language.
	LanguageScheme
	// LanguageSCSS represents the SCSS programming language.
	LanguageSCSS
	// LanguageSketchDrawing represents the SketchDrawing programming language.
	LanguageSketchDrawing
	// LanguageSKILL represents the SKILL programming language.
	LanguageSKILL
	// LanguageSlim represents the Slim programming language.
	LanguageSlim
	// LanguageSmali represents the Smali programming language.
	LanguageSmali
	// LanguageSmalltalk represents the Smalltalk programming language.
	LanguageSmalltalk
	// LanguageSourcePawn represents the SourcePawn programming language.
	LanguageSourcePawn
	// LanguageSQL represents the SQL programming language.
	LanguageSQL
	// LanguageSublimeTextConfig represents the SublimeTextConfig programming language.
	LanguageSublimeTextConfig
	// LanguageSvelte represents the Svelte programming language.
	LanguageSvelte
	// LanguageSwift represents the Swift programming language.
	LanguageSwift
	// LanguageSWIG represents the SWIG programming language.
	LanguageSWIG
	// Languagesystemverilog represents the systemverilog programming language.
	Languagesystemverilog
	// LanguageText represents the Text programming language.
	LanguageText
	// LanguageThrift represents the Thrift programming language.
	LanguageThrift
	// LanguageTOML represents the TOML programming language.
	LanguageTOML
	// LanguageTwig represents the Twig programming language.
	LanguageTwig
	// LanguageTypeScript represents the TypeScript programming language.
	LanguageTypeScript
	// LanguageTypoScript represents the TypoScript programming language.
	LanguageTypoScript
	// LanguageVB represents the VB programming language.
	LanguageVB
	// LanguageVBNet represents the VB.net programming language.
	LanguageVBNet
	// LanguageVCL represents the VCL programming language.
	LanguageVCL
	// LanguageVelocity represents the Velocity programming language.
	LanguageVelocity
	// LanguageVimL represents the VimL programming language.
	LanguageVimL
	// LanguageVueJS represents the VueJS programming language.
	LanguageVueJS
	// LanguageXAML represents the XAML programming language.
	LanguageXAML
	// LanguageXML represents the XML programming language.
	LanguageXML
	// LanguageXSLT represents the XSLT programming language.
	LanguageXSLT
	// LanguageYAML represents the YAML programming language.
	LanguageYAML
	// LanguageZig represents the Zig programming language.
	LanguageZig
)

const (
	languageUnkownStr            = "Unknown"
	languageAppleScriptStr       = "AppleScript"
	languageApacheConfStr        = "ApacheConf"
	languageAssemblyStr          = "Assembly"
	languageAwkStr               = "Awk"
	languageBashStr              = "Bash"
	languageBasicStr             = "Basic"
	languageCStr                 = "C"
	languageCPPStr               = "C++"
	languageCSharpStr            = "C#"
	languageClojureStr           = "Clojure"
	languageCMakeStr             = "CMake"
	languageCoffeeScriptStr      = "CoffeeScript"
	languageColdfusionHTMLStr    = "Coldfusion"
	languageCommonLispStr        = "Common Lisp"
	languageCrontabStr           = "Crontab"
	languageCrystalStr           = "Crystal"
	languageCSSStr               = "CSS"
	languageDartStr              = "Dart"
	languageDelphiStr            = "Delphi"
	languageDockerStr            = "Docker"
	languageElixirStr            = "Elixir"
	languageElmStr               = "Elm"
	languageEmacsLispStr         = "Emacs Lisp"
	languageErlangStr            = "Erlang"
	languageFSharpStr            = "F#"
	languageFortranStr           = "Fortran"
	languageGoStr                = "Go"
	languageGosuStr              = "Gosu"
	languageGroovyStr            = "Groovy"
	languageHAMLStr              = "HAML"
	languageHaskellStr           = "Haskell"
	languageHaxeStr              = "Haxe"
	languageHTMLStr              = "HTML"
	languageINIStr               = "INI"
	languageJavaStr              = "Java"
	languageJavaScriptStr        = "JavaScript"
	languageJSONStr              = "JSON"
	languageJSXStr               = "JSX"
	languageKotlinStr            = "Kotlin"
	languageLassoStr             = "Lasso"
	languageTexStr               = "TeX"
	languageLessStr              = "LESS"
	languageLiquidStr            = "liquid"
	languageLuaStr               = "Lua"
	languageMakefileStr          = "Makefile"
	languageMakoStr              = "Mako"
	languageMarkdownStr          = "Markdown"
	languageMarkoStr             = "Marko"
	languageMatlabStr            = "Matlab"
	languageModelicaStr          = "Modelica"
	languageModulaStr            = "Modula-2"
	languageMustacheStr          = "Mustache"
	languageNewLispStr           = "NewLisp"
	languageNixStr               = "Nix"
	languageObjectiveCStr        = "Objective-C"
	languageObjectiveCPPStr      = "Objective-C++"
	languageObjectiveJStr        = "Objective-J"
	languageOCamlStr             = "OCaml"
	languagePawnStr              = "Pawn"
	languagePerlStr              = "Perl"
	languagePHPStr               = "PHP"
	languagePostScriptStr        = "PostScript"
	languagePOVRayStr            = "POVRay"
	languagePowerShellStr        = "PowerShell"
	languagePrologStr            = "Prolog"
	languageProtocolBufferStr    = "Protocol Buffer"
	languagePugStr               = "Pug"
	languagePuppetStr            = "Puppet"
	languagePythonStr            = "Python"
	languageQMLStr               = "QML"
	languageRStr                 = "R"
	languageReasonMLStr          = "ReasonML"
	languageReStructuredTextStr  = "reStructuredText"
	languageRPMSpecStr           = "RPMSpec"
	languageRubyStr              = "Ruby"
	languageRustStr              = "Rust"
	languageSassStr              = "Sass"
	languageScalaStr             = "Scala"
	languageSchemeStr            = "Scheme"
	languageSCSSStr              = "SCSS"
	languageSketchDrawingStr     = "Sketch Drawing"
	languageSKILLStr             = "SKILL"
	languageSlimStr              = "Slim"
	languageSmaliStr             = "Smali"
	languageSmalltalkStr         = "Smalltalk"
	languageSourcePawnStr        = "SourcePawn"
	languageSQLStr               = "SQL"
	languageSublimeTextConfigStr = "Sublime Text Config"
	languageSvelteStr            = "Svelte"
	languageSwiftStr             = "Swift"
	languageSWIGStr              = "SWIG"
	languagesystemverilogStr     = "systemverilog"
	languageTextStr              = "Text"
	languageThriftStr            = "Thrift"
	languageTOMLStr              = "TOML"
	languageTwigStr              = "Twig"
	languageTypeScriptStr        = "TypeScript"
	languageTypoScriptStr        = "TypoScript"
	languageVBStr                = "VB"
	languageVBNetStr             = "VB.net"
	languageVCLStr               = "VCL"
	languageVelocityStr          = "Velocity"
	languageVimLStr              = "VimL"
	languageVueJSStr             = "Vue.js"
	languageXAMLStr              = "XAML"
	languageXMLStr               = "XML"
	languageXSLTStr              = "XSLT"
	languageYAMLStr              = "YAML"
	languageZigStr               = "Zig"
)

const (
	languageMakefileChromaStr  = "Base Makefile"
	languageFSharpChromaStr    = "FSharp"
	languageEmacsLispChromaStr = "EmacsLisp"
	languageAssemblyChromaStr  = "GAS"
	languageMarkdownChromaStr  = "markdown"
	languageTextChromaStr      = "plaintext"
)

// ParseLanguage parses a language from a string. Will return false
// as second parameter, if language could not be parsed.
// nolint:gocyclo
func ParseLanguage(s string) (Language, bool) {
	switch normalizeString(s) {
	case normalizeString(languageAppleScriptStr):
		return LanguageAppleScript, true
	case normalizeString(languageApacheConfStr):
		return LanguageApacheConf, true
	case normalizeString(languageAssemblyStr):
		return LanguageAssembly, true
	case normalizeString(languageAwkStr):
		return LanguageAwk, true
	case normalizeString(languageBasicStr):
		return LanguageBasic, true
	case normalizeString(languageBashStr):
		return LanguageBash, true
	case normalizeString(languageCStr):
		return LanguageC, true
	case normalizeString(languageCPPStr):
		return LanguageCPP, true
	case normalizeString(languageCSharpStr):
		return LanguageCSharp, true
	case normalizeString(languageClojureStr):
		return LanguageClojure, true
	case normalizeString(languageCMakeStr):
		return LanguageCMake, true
	case normalizeString(languageCoffeeScriptStr):
		return LanguageCoffeeScript, true
	case normalizeString(languageColdfusionHTMLStr):
		return LanguageColdfusionHTML, true
	case normalizeString(languageCommonLispStr):
		return LanguageCommonLisp, true
	case normalizeString(languageCrontabStr):
		return LanguageCrontab, true
	case normalizeString(languageCrystalStr):
		return LanguageCrystal, true
	case normalizeString(languageCSSStr):
		return LanguageCSS, true
	case normalizeString(languageDartStr):
		return LanguageDart, true
	case normalizeString(languageDelphiStr):
		return LanguageDelphi, true
	case normalizeString(languageDockerStr):
		return LanguageDocker, true
	case normalizeString(languageElixirStr):
		return LanguageElixir, true
	case normalizeString(languageElmStr):
		return LanguageElm, true
	case normalizeString(languageEmacsLispStr):
		return LanguageEmacsLisp, true
	case normalizeString(languageErlangStr):
		return LanguageErlang, true
	case normalizeString(languageFSharpStr):
		return LanguageFSharp, true
	case normalizeString(languageFortranStr):
		return LanguageFortran, true
	case normalizeString(languageGoStr):
		return LanguageGo, true
	case normalizeString(languageGosuStr):
		return LanguageGosu, true
	case normalizeString(languageGroovyStr):
		return LanguageGroovy, true
	case normalizeString(languageHAMLStr):
		return LanguageHAML, true
	case normalizeString(languageHaskellStr):
		return LanguageHaskell, true
	case normalizeString(languageHaxeStr):
		return LanguageHaxe, true
	case normalizeString(languageHTMLStr):
		return LanguageHTML, true
	case normalizeString(languageINIStr):
		return LanguageINI, true
	case normalizeString(languageJavaStr):
		return LanguageJava, true
	case normalizeString(languageJavaScriptStr):
		return LanguageJavaScript, true
	case normalizeString(languageJSONStr):
		return LanguageJSON, true
	case normalizeString(languageJSXStr):
		return LanguageJSX, true
	case normalizeString(languageKotlinStr):
		return LanguageKotlin, true
	case normalizeString(languageLassoStr):
		return LanguageLasso, true
	case normalizeString(languageTexStr):
		return LanguageTex, true
	case normalizeString(languageLessStr):
		return LanguageLess, true
	case normalizeString(languageLiquidStr):
		return LanguageLiquid, true
	case normalizeString(languageLuaStr):
		return LanguageLua, true
	case normalizeString(languageMakefileStr):
		return LanguageMakefile, true
	case normalizeString(languageMakoStr):
		return LanguageMako, true
	case normalizeString(languageMarkdownStr):
		return LanguageMarkdown, true
	case normalizeString(languageMarkoStr):
		return LanguageMarko, true
	case normalizeString(languageMatlabStr):
		return LanguageMatlab, true
	case normalizeString(languageModelicaStr):
		return LanguageModelica, true
	case normalizeString(languageModulaStr):
		return LanguageModula, true
	case normalizeString(languageMustacheStr):
		return LanguageMustache, true
	case normalizeString(languageNewLispStr):
		return LanguageNewLisp, true
	case normalizeString(languageNixStr):
		return LanguageNix, true
	case normalizeString(languageObjectiveCStr):
		return LanguageObjectiveC, true
	case normalizeString(languageObjectiveCPPStr):
		return LanguageObjectiveCPP, true
	case normalizeString(languageObjectiveJStr):
		return LanguageObjectiveJ, true
	case normalizeString(languageOCamlStr):
		return LanguageOCaml, true
	case normalizeString(languagePawnStr):
		return LanguagePawn, true
	case normalizeString(languagePerlStr):
		return LanguagePerl, true
	case normalizeString(languagePHPStr):
		return LanguagePHP, true
	case normalizeString(languagePostScriptStr):
		return LanguagePostScript, true
	case normalizeString(languagePOVRayStr):
		return LanguagePOVRay, true
	case normalizeString(languagePowerShellStr):
		return LanguagePowerShell, true
	case normalizeString(languagePrologStr):
		return LanguageProlog, true
	case normalizeString(languageProtocolBufferStr):
		return LanguageProtocolBuffer, true
	case normalizeString(languagePugStr):
		return LanguagePug, true
	case normalizeString(languagePuppetStr):
		return LanguagePuppet, true
	case normalizeString(languagePythonStr):
		return LanguagePython, true
	case normalizeString(languageQMLStr):
		return LanguageQML, true
	case normalizeString(languageRStr):
		return LanguageR, true
	case normalizeString(languageReasonMLStr):
		return LanguageReasonML, true
	case normalizeString(languageReStructuredTextStr):
		return LanguageReStructuredText, true
	case normalizeString(languageRPMSpecStr):
		return LanguageRPMSpec, true
	case normalizeString(languageRubyStr):
		return LanguageRuby, true
	case normalizeString(languageRustStr):
		return LanguageRust, true
	case normalizeString(languageSassStr):
		return LanguageSass, true
	case normalizeString(languageScalaStr):
		return LanguageScala, true
	case normalizeString(languageSchemeStr):
		return LanguageScheme, true
	case normalizeString(languageSCSSStr):
		return LanguageSCSS, true
	case normalizeString(languageSketchDrawingStr):
		return LanguageSketchDrawing, true
	case normalizeString(languageSKILLStr):
		return LanguageSKILL, true
	case normalizeString(languageSlimStr):
		return LanguageSlim, true
	case normalizeString(languageSmaliStr):
		return LanguageSmali, true
	case normalizeString(languageSmalltalkStr):
		return LanguageSmalltalk, true
	case normalizeString(languageSourcePawnStr):
		return LanguageSourcePawn, true
	case normalizeString(languageSQLStr):
		return LanguageSQL, true
	case normalizeString(languageSublimeTextConfigStr):
		return LanguageSublimeTextConfig, true
	case normalizeString(languageSvelteStr):
		return LanguageSvelte, true
	case normalizeString(languageSwiftStr):
		return LanguageSwift, true
	case normalizeString(languageSWIGStr):
		return LanguageSWIG, true
	case normalizeString(languagesystemverilogStr):
		return Languagesystemverilog, true
	case normalizeString(languageTextStr):
		return LanguageText, true
	case normalizeString(languageThriftStr):
		return LanguageThrift, true
	case normalizeString(languageTOMLStr):
		return LanguageTOML, true
	case normalizeString(languageTwigStr):
		return LanguageTwig, true
	case normalizeString(languageTypeScriptStr):
		return LanguageTypeScript, true
	case normalizeString(languageTypoScriptStr):
		return LanguageTypoScript, true
	case normalizeString(languageVBStr):
		return LanguageVB, true
	case normalizeString(languageVBNetStr):
		return LanguageVBNet, true
	case normalizeString(languageVCLStr):
		return LanguageVCL, true
	case normalizeString(languageVelocityStr):
		return LanguageVelocity, true
	case normalizeString(languageVimLStr):
		return LanguageVimL, true
	case normalizeString(languageVueJSStr):
		return LanguageVueJS, true
	case normalizeString(languageXAMLStr):
		return LanguageXAML, true
	case normalizeString(languageXMLStr):
		return LanguageXML, true
	case normalizeString(languageXSLTStr):
		return LanguageXSLT, true
	case normalizeString(languageYAMLStr):
		return LanguageYAML, true
	case normalizeString(languageZigStr):
		return LanguageZig, true
	default:
		return LanguageUnknown, false
	}
}

// ParseLanguageFromChroma parses a language from a chroma lexer name.
// Will return false as second parameter, if language could not be parsed.
// nolint:gocyclo
func ParseLanguageFromChroma(lexerName string) (Language, bool) {
	switch lexerName {
	case languageAssemblyChromaStr:
		return LanguageAssembly, true
	case languageEmacsLispChromaStr:
		return LanguageEmacsLisp, true
	case languageFSharpChromaStr:
		return LanguageFSharp, true
	case languageMakefileChromaStr:
		return LanguageMakefile, true
	case languageMarkdownChromaStr:
		return LanguageMarkdown, true
	case languageTextChromaStr:
		return LanguageText, true
	default:
		return ParseLanguage(lexerName)
	}
}

// MarshalJSON implements json.Marshaler interface.
func (l Language) MarshalJSON() ([]byte, error) {
	if l == LanguageUnknown {
		return []byte(`null`), nil
	}

	s := l.String()
	if s == "" {
		return nil, fmt.Errorf("invalid language %v", l)
	}

	return []byte(`"` + s + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (l *Language) UnmarshalJSON(v []byte) error {
	trimmed := strings.Trim(string(v), "\"")

	lang, _ := ParseLanguage(trimmed)

	*l = lang

	return nil
}

// String implements fmt.Stringer interface.
// nolint:gocyclo
func (l Language) String() string {
	switch l {
	case LanguageAppleScript:
		return languageAppleScriptStr
	case LanguageApacheConf:
		return languageApacheConfStr
	case LanguageAssembly:
		return languageAssemblyStr
	case LanguageAwk:
		return languageAwkStr
	case LanguageBasic:
		return languageBasicStr
	case LanguageBash:
		return languageBashStr
	case LanguageC:
		return languageCStr
	case LanguageCPP:
		return languageCPPStr
	case LanguageCSharp:
		return languageCSharpStr
	case LanguageClojure:
		return languageClojureStr
	case LanguageCMake:
		return languageCMakeStr
	case LanguageCoffeeScript:
		return languageCoffeeScriptStr
	case LanguageColdfusionHTML:
		return languageColdfusionHTMLStr
	case LanguageCommonLisp:
		return languageCommonLispStr
	case LanguageCrontab:
		return languageCrontabStr
	case LanguageCrystal:
		return languageCrystalStr
	case LanguageCSS:
		return languageCSSStr
	case LanguageDart:
		return languageDartStr
	case LanguageDelphi:
		return languageDelphiStr
	case LanguageDocker:
		return languageDockerStr
	case LanguageElixir:
		return languageElixirStr
	case LanguageElm:
		return languageElmStr
	case LanguageEmacsLisp:
		return languageEmacsLispStr
	case LanguageErlang:
		return languageErlangStr
	case LanguageFSharp:
		return languageFSharpStr
	case LanguageFortran:
		return languageFortranStr
	case LanguageGo:
		return languageGoStr
	case LanguageGosu:
		return languageGosuStr
	case LanguageGroovy:
		return languageGroovyStr
	case LanguageHAML:
		return languageHAMLStr
	case LanguageHaskell:
		return languageHaskellStr
	case LanguageHaxe:
		return languageHaxeStr
	case LanguageHTML:
		return languageHTMLStr
	case LanguageINI:
		return languageINIStr
	case LanguageJava:
		return languageJavaStr
	case LanguageJavaScript:
		return languageJavaScriptStr
	case LanguageJSON:
		return languageJSONStr
	case LanguageJSX:
		return languageJSXStr
	case LanguageKotlin:
		return languageKotlinStr
	case LanguageLasso:
		return languageLassoStr
	case LanguageTex:
		return languageTexStr
	case LanguageLess:
		return languageLessStr
	case LanguageLiquid:
		return languageLiquidStr
	case LanguageLua:
		return languageLuaStr
	case LanguageMakefile:
		return languageMakefileStr
	case LanguageMako:
		return languageMakoStr
	case LanguageMarkdown:
		return languageMarkdownStr
	case LanguageMarko:
		return languageMarkoStr
	case LanguageMatlab:
		return languageMatlabStr
	case LanguageModelica:
		return languageModelicaStr
	case LanguageModula:
		return languageModulaStr
	case LanguageMustache:
		return languageMustacheStr
	case LanguageNewLisp:
		return languageNewLispStr
	case LanguageNix:
		return languageNixStr
	case LanguageObjectiveC:
		return languageObjectiveCStr
	case LanguageObjectiveCPP:
		return languageObjectiveCPPStr
	case LanguageObjectiveJ:
		return languageObjectiveJStr
	case LanguageOCaml:
		return languageOCamlStr
	case LanguagePawn:
		return languagePawnStr
	case LanguagePerl:
		return languagePerlStr
	case LanguagePHP:
		return languagePHPStr
	case LanguagePostScript:
		return languagePostScriptStr
	case LanguagePOVRay:
		return languagePOVRayStr
	case LanguagePowerShell:
		return languagePowerShellStr
	case LanguageProlog:
		return languagePrologStr
	case LanguageProtocolBuffer:
		return languageProtocolBufferStr
	case LanguagePug:
		return languagePugStr
	case LanguagePuppet:
		return languagePuppetStr
	case LanguagePython:
		return languagePythonStr
	case LanguageQML:
		return languageQMLStr
	case LanguageR:
		return languageRStr
	case LanguageReasonML:
		return languageReasonMLStr
	case LanguageReStructuredText:
		return languageReStructuredTextStr
	case LanguageRPMSpec:
		return languageRPMSpecStr
	case LanguageRuby:
		return languageRubyStr
	case LanguageRust:
		return languageRustStr
	case LanguageSass:
		return languageSassStr
	case LanguageScala:
		return languageScalaStr
	case LanguageScheme:
		return languageSchemeStr
	case LanguageSCSS:
		return languageSCSSStr
	case LanguageSketchDrawing:
		return languageSketchDrawingStr
	case LanguageSKILL:
		return languageSKILLStr
	case LanguageSlim:
		return languageSlimStr
	case LanguageSmali:
		return languageSmaliStr
	case LanguageSmalltalk:
		return languageSmalltalkStr
	case LanguageSourcePawn:
		return languageSourcePawnStr
	case LanguageSQL:
		return languageSQLStr
	case LanguageSublimeTextConfig:
		return languageSublimeTextConfigStr
	case LanguageSvelte:
		return languageSvelteStr
	case LanguageSwift:
		return languageSwiftStr
	case LanguageSWIG:
		return languageSWIGStr
	case Languagesystemverilog:
		return languagesystemverilogStr
	case LanguageText:
		return languageTextStr
	case LanguageThrift:
		return languageThriftStr
	case LanguageTOML:
		return languageTOMLStr
	case LanguageTwig:
		return languageTwigStr
	case LanguageTypeScript:
		return languageTypeScriptStr
	case LanguageTypoScript:
		return languageTypoScriptStr
	case LanguageVB:
		return languageVBStr
	case LanguageVBNet:
		return languageVBNetStr
	case LanguageVCL:
		return languageVCLStr
	case LanguageVelocity:
		return languageVelocityStr
	case LanguageVimL:
		return languageVimLStr
	case LanguageVueJS:
		return languageVueJSStr
	case LanguageXAML:
		return languageXAMLStr
	case LanguageXML:
		return languageXMLStr
	case LanguageXSLT:
		return languageXSLTStr
	case LanguageYAML:
		return languageYAMLStr
	case LanguageZig:
		return languageZigStr
	default:
		return languageUnkownStr
	}
}

// StringChroma returns the corresponding chroma lexer name.
// nolint:gocyclo
func (l Language) StringChroma() string {
	switch l {
	case LanguageAssembly:
		return languageAssemblyChromaStr
	case LanguageEmacsLisp:
		return languageEmacsLispChromaStr
	case LanguageFSharp:
		return languageFSharpChromaStr
	case LanguageMakefile:
		return languageMakefileChromaStr
	case LanguageMarkdown:
		return languageMarkdownChromaStr
	case LanguageText:
		return languageTextChromaStr
	default:
		return l.String()
	}
}

func normalizeString(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	s = strings.Replace(s, " ", "", -1)
	s = strings.Replace(s, "-", "", -1)
	s = strings.Replace(s, "#", "sharp", -1)
	return strings.Replace(s, "++", "pp", -1)
}
