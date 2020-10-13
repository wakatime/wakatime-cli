package heartbeat

import (
	"fmt"
	"strings"

	jww "github.com/spf13/jwalterweatherman"
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
	// LanguageMako represents the Mako programming language.
	LanguageMako
	// LanguageMarkdown represents the Markdown programming language.
	LanguageMarkdown
	// LanguageMarko represents the Marko programming language.
	LanguageMarko
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
	// LanguageVB represents the VB programming language.
	LanguageVB
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
	languageMakoStr              = "Mako"
	languageMarkdownStr          = "Markdown"
	languageMarkoStr             = "Marko"
	languageModelicaStr          = "Modelica"
	languageModulaStr            = "Modula-2"
	languageMustacheStr          = "Mustache"
	languageNewLispStr           = "NewLisp"
	languageNixStr               = "Nix"
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
	languageVBStr                = "VB.net"
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
	languageFSharpChromaStr    = "F#"
	languageEmacsLispChromaStr = "EmacsLisp"
	languageAssemblyChromaStr  = "GAS"
	languageMarkdownChromaStr  = "markdown"
	languageTextChromaStr      = "plaintext"
)

// ParseLanguage parses a language from a string.
// nolint:gocyclo
func ParseLanguage(s string) Language {
	switch s {
	case languageAppleScriptStr:
		return LanguageAppleScript
	case languageApacheConfStr:
		return LanguageApacheConf
	case languageAssemblyStr:
		return LanguageAssembly
	case languageAwkStr:
		return LanguageAwk
	case languageBashStr:
		return LanguageBash
	case languageCStr:
		return LanguageC
	case languageCPPStr:
		return LanguageCPP
	case languageCSharpStr:
		return LanguageCSharp
	case languageClojureStr:
		return LanguageClojure
	case languageCMakeStr:
		return LanguageCMake
	case languageCoffeeScriptStr:
		return LanguageCoffeeScript
	case languageColdfusionHTMLStr:
		return LanguageColdfusionHTML
	case languageCommonLispStr:
		return LanguageCommonLisp
	case languageCrontabStr:
		return LanguageCrontab
	case languageCrystalStr:
		return LanguageCrystal
	case languageCSSStr:
		return LanguageCSS
	case languageDartStr:
		return LanguageDart
	case languageDelphiStr:
		return LanguageDelphi
	case languageDockerStr:
		return LanguageDocker
	case languageElixirStr:
		return LanguageElixir
	case languageElmStr:
		return LanguageElm
	case languageEmacsLispStr:
		return LanguageEmacsLisp
	case languageErlangStr:
		return LanguageErlang
	case languageFSharpStr:
		return LanguageFSharp
	case languageFortranStr:
		return LanguageFortran
	case languageGoStr:
		return LanguageGo
	case languageGosuStr:
		return LanguageGosu
	case languageGroovyStr:
		return LanguageGroovy
	case languageHaskellStr:
		return LanguageHaskell
	case languageHaxeStr:
		return LanguageHaxe
	case languageHTMLStr:
		return LanguageHTML
	case languageINIStr:
		return LanguageINI
	case languageJavaStr:
		return LanguageJava
	case languageJavaScriptStr:
		return LanguageJavaScript
	case languageJSONStr:
		return LanguageJSON
	case languageJSXStr:
		return LanguageJSX
	case languageKotlinStr:
		return LanguageKotlin
	case languageLassoStr:
		return LanguageLasso
	case languageTexStr:
		return LanguageTex
	case languageLessStr:
		return LanguageLess
	case languageLiquidStr:
		return LanguageLiquid
	case languageLuaStr:
		return LanguageLua
	case languageMakoStr:
		return LanguageMako
	case languageMarkdownStr:
		return LanguageMarkdown
	case languageMarkoStr:
		return LanguageMarko
	case languageModelicaStr:
		return LanguageModelica
	case languageModulaStr:
		return LanguageModula
	case languageMustacheStr:
		return LanguageMustache
	case languageNewLispStr:
		return LanguageNewLisp
	case languageNixStr:
		return LanguageNix
	case languageObjectiveJStr:
		return LanguageObjectiveJ
	case languageOCamlStr:
		return LanguageOCaml
	case languagePawnStr:
		return LanguagePawn
	case languagePerlStr:
		return LanguagePerl
	case languagePHPStr:
		return LanguagePHP
	case languagePostScriptStr:
		return LanguagePostScript
	case languagePOVRayStr:
		return LanguagePOVRay
	case languagePowerShellStr:
		return LanguagePowerShell
	case languagePrologStr:
		return LanguageProlog
	case languageProtocolBufferStr:
		return LanguageProtocolBuffer
	case languagePugStr:
		return LanguagePug
	case languagePuppetStr:
		return LanguagePuppet
	case languagePythonStr:
		return LanguagePython
	case languageQMLStr:
		return LanguageQML
	case languageRStr:
		return LanguageR
	case languageReasonMLStr:
		return LanguageReasonML
	case languageReStructuredTextStr:
		return LanguageReStructuredText
	case languageRPMSpecStr:
		return LanguageRPMSpec
	case languageRubyStr:
		return LanguageRuby
	case languageRustStr:
		return LanguageRust
	case languageSassStr:
		return LanguageSass
	case languageScalaStr:
		return LanguageScala
	case languageSchemeStr:
		return LanguageScheme
	case languageSCSSStr:
		return LanguageSCSS
	case languageSketchDrawingStr:
		return LanguageSketchDrawing
	case languageSlimStr:
		return LanguageSlim
	case languageSmaliStr:
		return LanguageSmali
	case languageSmalltalkStr:
		return LanguageSmalltalk
	case languageSourcePawnStr:
		return LanguageSourcePawn
	case languageSQLStr:
		return LanguageSQL
	case languageSublimeTextConfigStr:
		return LanguageSublimeTextConfig
	case languageSvelteStr:
		return LanguageSvelte
	case languageSwiftStr:
		return LanguageSwift
	case languageSWIGStr:
		return LanguageSWIG
	case languagesystemverilogStr:
		return Languagesystemverilog
	case languageTextStr:
		return LanguageText
	case languageThriftStr:
		return LanguageThrift
	case languageTOMLStr:
		return LanguageTOML
	case languageTwigStr:
		return LanguageTwig
	case languageTypeScriptStr:
		return LanguageTypeScript
	case languageVBStr:
		return LanguageVB
	case languageVCLStr:
		return LanguageVCL
	case languageVelocityStr:
		return LanguageVelocity
	case languageVimLStr:
		return LanguageVimL
	case languageVueJSStr:
		return LanguageVueJS
	case languageXAMLStr:
		return LanguageXAML
	case languageXMLStr:
		return LanguageXML
	case languageXSLTStr:
		return LanguageXSLT
	case languageYAMLStr:
		return LanguageYAML
	case languageZigStr:
		return LanguageZig
	default:
		return LanguageUnknown
	}
}

// ParseLanguageFromChroma parses a language from a chroma lexer name.
// nolint:gocyclo
func ParseLanguageFromChroma(lexerName string) Language {
	switch lexerName {
	case languageAppleScriptStr:
		return LanguageAppleScript
	case languageApacheConfStr:
		return LanguageApacheConf
	case languageAssemblyChromaStr:
		return LanguageAssembly
	case languageAwkStr:
		return LanguageAwk
	case languageBashStr:
		return LanguageBash
	case languageCStr:
		return LanguageC
	case languageCPPStr:
		return LanguageCPP
	case languageCSharpStr:
		return LanguageCSharp
	case languageClojureStr:
		return LanguageClojure
	case languageCMakeStr:
		return LanguageCMake
	case languageCoffeeScriptStr:
		return LanguageCoffeeScript
	case languageCommonLispStr:
		return LanguageCommonLisp
	case languageCrystalStr:
		return LanguageCrystal
	case languageCSSStr:
		return LanguageCSS
	case languageDartStr:
		return LanguageDart
	case languageDockerStr:
		return LanguageDocker
	case languageElixirStr:
		return LanguageElixir
	case languageElmStr:
		return LanguageElm
	case languageEmacsLispChromaStr:
		return LanguageEmacsLisp
	case languageErlangStr:
		return LanguageErlang
	case languageFSharpChromaStr:
		return LanguageFSharp
	case languageFortranStr:
		return LanguageFortran
	case languageGoStr:
		return LanguageGo
	case languageGroovyStr:
		return LanguageGroovy
	case languageHaskellStr:
		return LanguageHaskell
	case languageHaxeStr:
		return LanguageHaxe
	case languageHTMLStr:
		return LanguageHTML
	case languageINIStr:
		return LanguageINI
	case languageJavaStr:
		return LanguageJava
	case languageJavaScriptStr:
		return LanguageJavaScript
	case languageJSONStr:
		return LanguageJSON
	case languageKotlinStr:
		return LanguageKotlin
	case languageTexStr:
		return LanguageTex
	case languageLuaStr:
		return LanguageLua
	case languageMakoStr:
		return LanguageMako
	case languageMarkdownChromaStr:
		return LanguageMarkdown
	case languageModulaStr:
		return LanguageModula
	case languageNixStr:
		return LanguageNix
	case languageOCamlStr:
		return LanguageOCaml
	case languagePerlStr:
		return LanguagePerl
	case languagePHPStr:
		return LanguagePHP
	case languagePostScriptStr:
		return LanguagePostScript
	case languagePOVRayStr:
		return LanguagePOVRay
	case languagePowerShellStr:
		return LanguagePowerShell
	case languagePrologStr:
		return LanguageProlog
	case languageProtocolBufferStr:
		return LanguageProtocolBuffer
	case languagePuppetStr:
		return LanguagePuppet
	case languagePythonStr:
		return LanguagePython
	case languageRStr:
		return LanguageR
	case languageReasonMLStr:
		return LanguageReasonML
	case languageReStructuredTextStr:
		return LanguageReStructuredText
	case languageRubyStr:
		return LanguageRuby
	case languageRustStr:
		return LanguageRust
	case languageSassStr:
		return LanguageSass
	case languageScalaStr:
		return LanguageScala
	case languageSchemeStr:
		return LanguageScheme
	case languageSCSSStr:
		return LanguageSCSS
	case languageSmalltalkStr:
		return LanguageSmalltalk
	case languageSQLStr:
		return LanguageSQL
	case languageSwiftStr:
		return LanguageSwift
	case languagesystemverilogStr:
		return Languagesystemverilog
	case languageTextChromaStr:
		return LanguageText
	case languageThriftStr:
		return LanguageThrift
	case languageTOMLStr:
		return LanguageTOML
	case languageTwigStr:
		return LanguageTwig
	case languageTypeScriptStr:
		return LanguageTypeScript
	case languageVBStr:
		return LanguageVB
	case languageVimLStr:
		return LanguageVimL
	case languageXMLStr:
		return LanguageXML
	case languageYAMLStr:
		return LanguageYAML
	case languageZigStr:
		return LanguageZig
	default:
		jww.WARN.Printf("Could not detect language. Unknown lexer name: %q", lexerName)
		return LanguageUnknown
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

	lang := ParseLanguage(trimmed)

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
	case LanguageMako:
		return languageMakoStr
	case LanguageMarkdown:
		return languageMarkdownStr
	case LanguageMarko:
		return languageMarkoStr
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
	case LanguageVB:
		return languageVBStr
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
